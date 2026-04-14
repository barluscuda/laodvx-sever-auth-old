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

func (r *cachedTenantRepository) GetByTenantName(tenantName string) (*model.Tenant, error) {
	if t, found, cached := r.get(tenantName); cached {
		if !found {
			return nil, ports.ErrNotFound
		}
		return t, nil
	}
	t, err := r.repo.GetByTenantName(tenantName)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.setNotFound(tenantName)
		}
		return nil, err
	}
	r.set(t)
	return t, nil
}

func (r *cachedTenantRepository) Update(t *model.Tenant) error {
	if err := r.repo.Update(t); err != nil {
		return err
	}
	r.set(t)
	return nil
}

func (r *cachedTenantRepository) Delete(tenantName string) error {
	if err := r.repo.Delete(tenantName); err != nil {
		return err
	}
	r.rdb.Del(context.Background(), tenantCacheKey(tenantName))
	return nil
}

func (r *cachedTenantRepository) ExistsByName(tenantName string) (bool, error) {
	if _, found, cached := r.get(tenantName); cached {
		return found, nil
	}
	t, err := r.repo.GetByTenantName(tenantName)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.setNotFound(tenantName)
			return false, nil
		}
		return false, err
	}
	r.set(t)
	return true, nil
}

func (r *cachedTenantRepository) ExistsByUUID(id uuid.UUID) (bool, error) {
	return r.repo.ExistsByUUID(id)
}

func (r *cachedTenantRepository) Ban(tenantName string) error {
	if err := r.repo.Ban(tenantName); err != nil {
		return err
	}
	r.rdb.Del(context.Background(), tenantCacheKey(tenantName))
	return nil
}

func (r *cachedTenantRepository) set(t *model.Tenant) {
	r.setEntry(t.TenantName, cachedTenantEntry{Found: true, Tenant: t})
}

func (r *cachedTenantRepository) setNotFound(tenantName string) {
	r.setEntry(tenantName, cachedTenantEntry{Found: false})
}

func (r *cachedTenantRepository) setEntry(tenantName string, entry cachedTenantEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	r.rdb.Set(context.Background(), tenantCacheKey(tenantName), data, r.ttl)
}

func (r *cachedTenantRepository) get(tenantName string) (*model.Tenant, bool, bool) {
	data, err := r.rdb.Get(context.Background(), tenantCacheKey(tenantName)).Bytes()
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

func tenantCacheKey(tenantName string) string {
	return fmt.Sprintf("tenant:%s", tenantName)
}
