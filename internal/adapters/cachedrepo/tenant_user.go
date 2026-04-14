package cachedrepo

import (
	"errors"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/cache"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
)

// TenantUser is the cache-aside coordinator for tenant users: cache first,
// DB fallback on miss, then write back.
type TenantUser struct {
	repo  ports.TenantUserRepository
	cache *cache.TenantUserCache
}

func NewTenantUser(repo ports.TenantUserRepository, c *cache.TenantUserCache) ports.TenantUserRepository {
	return &TenantUser{repo: repo, cache: c}
}

func (r *TenantUser) Create(u *model.TenantUser) error {
	if err := r.repo.Create(u); err != nil {
		return err
	}
	r.cache.Set(u)
	return nil
}

func (r *TenantUser) GetByID(tenantID, id uuid.UUID) (*model.TenantUser, error) {
	if u, found, hit := r.cache.GetByID(tenantID, id); hit {
		if !found {
			return nil, ports.ErrNotFound
		}
		return u, nil
	}
	u, err := r.repo.GetByID(tenantID, id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.cache.SetNotFoundByID(tenantID, id)
		}
		return nil, err
	}
	r.cache.Set(u)
	return u, nil
}

func (r *TenantUser) GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	if u, found, hit := r.cache.GetByEmail(tenantID, email); hit {
		if !found {
			return nil, ports.ErrNotFound
		}
		return u, nil
	}
	u, err := r.repo.GetByEmail(tenantID, email)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.cache.SetNotFoundByEmail(tenantID, email)
		}
		return nil, err
	}
	r.cache.Set(u)
	return u, nil
}

func (r *TenantUser) GetAll(tenantID uuid.UUID) ([]model.TenantUser, error) {
	return r.repo.GetAll(tenantID)
}
