package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)


type cachedTenantUser struct {
	UUID          uuid.UUID  `json:"uuid"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	Email         string     `json:"email"`
	Role          string     `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type cachedTenantUserEntry struct {
	Found      bool              `json:"found"`
	TenantUser *cachedTenantUser `json:"user,omitempty"`
}

type cachedAuthTenantUser struct {
	UUID          uuid.UUID  `json:"uuid"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	Email         string     `json:"email"`
	Password      string     `json:"password"`
	Role          string     `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type cachedAuthTenantUserEntry struct {
	Found      bool                  `json:"found"`
	TenantUser *cachedAuthTenantUser `json:"user,omitempty"`
}

type cachedTenantUserRepository struct {
	repo ports.TenantUserRepository
	rdb  *redis.Client
	ttl  time.Duration
}

func NewCachedTenantUserRepository(repo ports.TenantUserRepository, rdb *redis.Client, ttl time.Duration) ports.TenantUserRepository {
	return &cachedTenantUserRepository{repo: repo, rdb: rdb, ttl: ttl}
}

func (r *cachedTenantUserRepository) Create(u *model.TenantUser) error {
	if err := r.repo.Create(u); err != nil {
		return err
	}
	r.setByID(u.TenantID, u)
	r.setByEmail(u.TenantID, u)
	return nil
}

func (r *cachedTenantUserRepository) GetByID(tenantID, id uuid.UUID) (*model.TenantUser, error) {
	if u, found, cached := r.get(tenantID, id); cached {
		if !found {
			return nil, ports.ErrNotFound
		}
		return u, nil
	}
	u, err := r.repo.GetByID(tenantID, id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.setNotFoundByID(tenantID, id)
		}
		return nil, err
	}
	r.setByID(tenantID, u)
	r.setByEmail(tenantID, u)
	return u, nil
}

func (r *cachedTenantUserRepository) GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	if u, found, cached := r.getByEmail(tenantID, email); cached {
		if !found {
			return nil, ports.ErrNotFound
		}
		return u, nil
	}
	u, err := r.repo.GetByEmail(tenantID, email)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.setNotFoundByEmail(tenantID, email)
		}
		return nil, err
	}
	r.setByID(tenantID, u)
	r.setByEmail(tenantID, u)
	return u, nil
}

func (r *cachedTenantUserRepository) GetAll(tenantID uuid.UUID) ([]model.TenantUser, error) {
	return r.repo.GetAll(tenantID)
}

func (r *cachedTenantUserRepository) SystemGetAll() ([]model.TenantUser, error) {
	return r.repo.SystemGetAll()
}

func (r *cachedTenantUserRepository) SystemGetAllByEmail(email string) ([]model.TenantUser, error) {
	return r.repo.SystemGetAllByEmail(email)
}

func (r *cachedTenantUserRepository) SystemGetByID(id uuid.UUID) (*model.TenantUser, error) {
	if u, found, cached := r.get(uuid.Nil, id); cached {
		if !found {
			return nil, ports.ErrNotFound
		}
		return u, nil
	}
	u, err := r.repo.SystemGetByID(id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.setNotFoundByID(uuid.Nil, id)
		}
		return nil, err
	}
	r.setByID(uuid.Nil, u)
	return u, nil
}

func (r *cachedTenantUserRepository) setByID(tenantID uuid.UUID, u *model.TenantUser) {
	r.setIDEntry(tenantID, u.UUID, cachedTenantUserEntry{
		Found: true,
		TenantUser: &cachedTenantUser{
			UUID:          u.UUID,
			TenantID:      u.TenantID,
			Email:         u.Email,
			Role:          u.Role,
			EmailVerified: u.EmailVerified,
			CreatedAt:     u.CreatedAt,
			UpdatedAt:     u.UpdatedAt,
		},
	})
}

func (r *cachedTenantUserRepository) setNotFoundByID(tenantID, id uuid.UUID) {
	r.setIDEntry(tenantID, id, cachedTenantUserEntry{Found: false})
}

func (r *cachedTenantUserRepository) setIDEntry(tenantID, id uuid.UUID, entry cachedTenantUserEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	r.rdb.Set(context.Background(), tenantUserIDCacheKey(tenantID, id), data, r.ttl)
}

func (r *cachedTenantUserRepository) setByEmail(tenantID uuid.UUID, u *model.TenantUser) {
	r.setEmailEntry(tenantID, u.Email, cachedAuthTenantUserEntry{
		Found: true,
		TenantUser: &cachedAuthTenantUser{
			UUID:          u.UUID,
			TenantID:      u.TenantID,
			Email:         u.Email,
			Password:      u.Password,
			Role:          u.Role,
			EmailVerified: u.EmailVerified,
			CreatedAt:     u.CreatedAt,
			UpdatedAt:     u.UpdatedAt,
		},
	})
}

func (r *cachedTenantUserRepository) setNotFoundByEmail(tenantID uuid.UUID, email string) {
	r.setEmailEntry(tenantID, email, cachedAuthTenantUserEntry{Found: false})
}

func (r *cachedTenantUserRepository) setEmailEntry(tenantID uuid.UUID, email string, entry cachedAuthTenantUserEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	r.rdb.Set(context.Background(), tenantUserEmailCacheKey(tenantID, email), data, r.ttl)
}

func (r *cachedTenantUserRepository) get(tenantID, id uuid.UUID) (*model.TenantUser, bool, bool) {
	data, err := r.rdb.Get(context.Background(), tenantUserIDCacheKey(tenantID, id)).Bytes()
	if err != nil {
		return nil, false, false
	}

	var entry cachedTenantUserEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, false
	}
	if !entry.Found {
		return nil, false, true
	}
	if entry.TenantUser == nil {
		return nil, false, false
	}

	cu := entry.TenantUser
	return &model.TenantUser{
		UUID:          cu.UUID,
		TenantID:      cu.TenantID,
		Email:         cu.Email,
		Role:          cu.Role,
		EmailVerified: cu.EmailVerified,
		CreatedAt:     cu.CreatedAt,
		UpdatedAt:     cu.UpdatedAt,
	}, true, true
}

func (r *cachedTenantUserRepository) getByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, bool, bool) {
	data, err := r.rdb.Get(context.Background(), tenantUserEmailCacheKey(tenantID, email)).Bytes()
	if err != nil {
		return nil, false, false
	}

	var entry cachedAuthTenantUserEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, false
	}
	if !entry.Found {
		return nil, false, true
	}
	if entry.TenantUser == nil {
		return nil, false, false
	}

	cu := entry.TenantUser
	return &model.TenantUser{
		UUID:          cu.UUID,
		TenantID:      cu.TenantID,
		Email:         cu.Email,
		Password:      cu.Password,
		Role:          cu.Role,
		EmailVerified: cu.EmailVerified,
		CreatedAt:     cu.CreatedAt,
		UpdatedAt:     cu.UpdatedAt,
	}, true, true
}


func tenantUserIDCacheKey(tenantID, id uuid.UUID) string {
	return fmt.Sprintf("user:%s:%s", tenantID, id)
}

func tenantUserEmailCacheKey(tenantID uuid.UUID, email string) string {
	return fmt.Sprintf("user-email:%s:%s", tenantID, strings.ToLower(strings.TrimSpace(email)))
}
