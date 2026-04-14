package cachedrepo

import (
	"errors"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/cache"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
)

// Tenant is the cache-aside coordinator: it checks the cache first and falls
// back to the underlying DB repository on miss, writing the result back.
type Tenant struct {
	repo  ports.TenantRepository
	cache *cache.TenantCache
}

func NewTenant(repo ports.TenantRepository, c *cache.TenantCache) ports.TenantRepository {
	return &Tenant{repo: repo, cache: c}
}

func (r *Tenant) Create(t *model.Tenant) error {
	if err := r.repo.Create(t); err != nil {
		return err
	}
	r.cache.Set(t)
	return nil
}

func (r *Tenant) GetAll() ([]model.Tenant, error) {
	return r.repo.GetAll()
}

func (r *Tenant) GetByTenantName(tenantName string) (*model.Tenant, error) {
	if t, found, hit := r.cache.Get(tenantName); hit {
		if !found {
			return nil, ports.ErrNotFound
		}
		return t, nil
	}
	t, err := r.repo.GetByTenantName(tenantName)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.cache.SetNotFound(tenantName)
		}
		return nil, err
	}
	r.cache.Set(t)
	return t, nil
}

func (r *Tenant) Update(t *model.Tenant) error {
	if err := r.repo.Update(t); err != nil {
		return err
	}
	r.cache.Set(t)
	return nil
}

func (r *Tenant) Delete(tenantName string) error {
	if err := r.repo.Delete(tenantName); err != nil {
		return err
	}
	r.cache.Delete(tenantName)
	return nil
}

func (r *Tenant) ExistsByName(tenantName string) (bool, error) {
	if _, found, hit := r.cache.Get(tenantName); hit {
		return found, nil
	}
	return r.repo.ExistsByName(tenantName)
}

func (r *Tenant) ExistsByUUID(id uuid.UUID) (bool, error) {
	return r.repo.ExistsByUUID(id)
}

func (r *Tenant) Ban(tenantName string) error {
	if err := r.repo.Ban(tenantName); err != nil {
		return err
	}
	r.cache.Delete(tenantName)
	return nil
}
