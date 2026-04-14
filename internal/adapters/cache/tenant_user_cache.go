package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type tenantUserDTO struct {
	ID        uint64    `json:"id"`
	UUID      uuid.UUID `json:"uuid"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantUserCache struct {
	store *Store[tenantUserDTO]
}

func NewTenantUserCache(rdb *redis.Client, ttl time.Duration) *TenantUserCache {
	return &TenantUserCache{store: NewStore[tenantUserDTO](rdb, ttl)}
}

func (c *TenantUserCache) GetByID(tenantID, id uuid.UUID) (u *model.TenantUser, found, hit bool) {
	return c.getByKey(tenantUserIDKey(tenantID, id))
}

func (c *TenantUserCache) GetByEmail(tenantID uuid.UUID, email string) (u *model.TenantUser, found, hit bool) {
	return c.getByKey(tenantUserEmailKey(tenantID, email))
}

// Set writes both the ID-keyed and email-keyed entries for the user.
func (c *TenantUserCache) Set(u *model.TenantUser) {
	dto := toTenantUserDTO(u)
	ctx := context.Background()
	c.store.Set(ctx, tenantUserIDKey(u.TenantID, u.UUID), dto)
	c.store.Set(ctx, tenantUserEmailKey(u.TenantID, u.Email), dto)
}

func (c *TenantUserCache) SetNotFoundByID(tenantID, id uuid.UUID) {
	c.store.SetNotFound(context.Background(), tenantUserIDKey(tenantID, id))
}

func (c *TenantUserCache) SetNotFoundByEmail(tenantID uuid.UUID, email string) {
	c.store.SetNotFound(context.Background(), tenantUserEmailKey(tenantID, email))
}

func (c *TenantUserCache) getByKey(key string) (*model.TenantUser, bool, bool) {
	dto, f, h := c.store.Get(context.Background(), key)
	if !h {
		return nil, false, false
	}
	if !f {
		return nil, false, true
	}
	return fromTenantUserDTO(dto), true, true
}

func toTenantUserDTO(u *model.TenantUser) *tenantUserDTO {
	return &tenantUserDTO{
		ID:        u.ID,
		UUID:      u.UUID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		Password:  u.Password,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func fromTenantUserDTO(d *tenantUserDTO) *model.TenantUser {
	return &model.TenantUser{
		ID:        d.ID,
		UUID:      d.UUID,
		TenantID:  d.TenantID,
		Email:     d.Email,
		Password:  d.Password,
		Role:      d.Role,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func tenantUserIDKey(tenantID, id uuid.UUID) string {
	return fmt.Sprintf("tenant-user:%s:%s", tenantID, id)
}

func tenantUserEmailKey(tenantID uuid.UUID, email string) string {
	return fmt.Sprintf("tenant-user-email:%s:%s", tenantID, strings.ToLower(strings.TrimSpace(email)))
}
