package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type cachedTenantRepository struct {
	repo ports.TenantRepository
	rdb  *redis.Client
	ttl  time.Duration
}

type cachedTenantEntry struct {
	Found  bool          `json:"found"`
	Tenant *model.Tenant `json:"tenant,omitempty"`
}

func NewCachedTenantRepository(repo ports.TenantRepository, rdb *redis.Client, ttl time.Duration) ports.TenantRepository {
	return &cachedTenantRepository{repo: repo, rdb: rdb, ttl: ttl}
}

func (r *cachedTenantRepository) Create(t *model.Tenant) error {
	if err := r.repo.Create(t); err != nil {
		return err
	}
	r.set(t)
	return nil
}

func (r *cachedTenantRepository) GetAll() ([]model.Tenant, error) {
	return r.repo.GetAll()
}

func (r *cachedTenantRepository) GetByID(id uuid.UUID) (*model.Tenant, error) {
	if t, found, cached := r.get(id); cached {
		if !found {
			return nil, ports.ErrNotFound
		}
		return t, nil
	}
	t, err := r.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.setNotFound(id)
		}
		return nil, err
	}
	r.set(t)
	return t, nil
}

func (r *cachedTenantRepository) GetByName(name string) (*model.Tenant, error) {
	return r.repo.GetByName(name)
}

func (r *cachedTenantRepository) Update(t *model.Tenant) error {
	if err := r.repo.Update(t); err != nil {
		return err
	}
	r.set(t)
	return nil
}

func (r *cachedTenantRepository) Delete(id uuid.UUID) error {
	if err := r.repo.Delete(id); err != nil {
		return err
	}
	r.rdb.Del(context.Background(), tenantCacheKey(id))
	return nil
}

func (r *cachedTenantRepository) ExistsByID(id uuid.UUID) (bool, error) {
	if _, found, cached := r.get(id); cached {
		return found, nil
	}
	t, err := r.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.setNotFound(id)
			return false, nil
		}
		return false, err
	}
	r.set(t)
	return true, nil
}

func (r *cachedTenantRepository) set(t *model.Tenant) {
	r.setEntry(t.UUID, cachedTenantEntry{Found: true, Tenant: t})
}

func (r *cachedTenantRepository) setNotFound(id uuid.UUID) {
	r.setEntry(id, cachedTenantEntry{Found: false})
}

func (r *cachedTenantRepository) setEntry(id uuid.UUID, entry cachedTenantEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	r.rdb.Set(context.Background(), tenantCacheKey(id), data, r.ttl)
}

func (r *cachedTenantRepository) get(id uuid.UUID) (*model.Tenant, bool, bool) {
	data, err := r.rdb.Get(context.Background(), tenantCacheKey(id)).Bytes()
	if err != nil {
		return nil, false, false
	}

	var entry cachedTenantEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, false
	}
	if !entry.Found {
		return nil, false, true
	}
	if entry.Tenant == nil {
		return nil, false, false
	}
	return entry.Tenant, true, true
}

func tenantCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("tenant:%s", id)
}
