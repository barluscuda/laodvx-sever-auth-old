package dto

import (
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
)

type CreateTenantRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type UpdateTenantRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type TenantResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToTenantResponse(t *model.Tenant) TenantResponse {
	return TenantResponse{
		ID:        t.UUID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func ToTenantResponseList(tenants []model.Tenant) []TenantResponse {
	res := make([]TenantResponse, len(tenants))
	for i := range tenants {
		res[i] = ToTenantResponse(&tenants[i])
	}
	return res
}
