package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type tenantDTO struct {
	ID         uint64    `json:"id"`
	UUID       uuid.UUID `json:"uuid"`
	TenantName string    `json:"tenant_name"`
	Label      string    `json:"label"`
	Plan       string    `json:"plan"`
	Banned     bool      `json:"banned"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type TenantCache struct {
	store *Store[tenantDTO]
}

func NewTenantCache(rdb *redis.Client, ttl time.Duration) *TenantCache {
	return &TenantCache{store: NewStore[tenantDTO](rdb, ttl)}
}

func (c *TenantCache) Get(tenantName string) (t *model.Tenant, found, hit bool) {
	dto, f, h := c.store.Get(context.Background(), tenantKey(tenantName))
	if !h {
		return nil, false, false
	}
	if !f {
		return nil, false, true
	}
	return fromTenantDTO(dto), true, true
}

func (c *TenantCache) Set(t *model.Tenant) {
	c.store.Set(context.Background(), tenantKey(t.TenantName), toTenantDTO(t))
}

func (c *TenantCache) SetNotFound(tenantName string) {
	c.store.SetNotFound(context.Background(), tenantKey(tenantName))
}

func (c *TenantCache) Delete(tenantName string) {
	c.store.Delete(context.Background(), tenantKey(tenantName))
}

func toTenantDTO(t *model.Tenant) *tenantDTO {
	return &tenantDTO{
		ID:         t.ID,
		UUID:       t.UUID,
		TenantName: t.TenantName,
		Label:      t.Label,
		Plan:       t.Plan,
		Banned:     t.Banned,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}
}

func fromTenantDTO(d *tenantDTO) *model.Tenant {
	return &model.Tenant{
		ID:         d.ID,
		UUID:       d.UUID,
		TenantName: d.TenantName,
		Label:      d.Label,
		Plan:       d.Plan,
		Banned:     d.Banned,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}

func tenantKey(tenantName string) string {
	return fmt.Sprintf("tenant:%s", tenantName)
}
