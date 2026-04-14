package dto

import (
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
)

type CreateTenantRequest struct {
	TenantName string `json:"tenant_name" binding:"required,min=1,max=100,alphanum"`
	Label      string `json:"label"       binding:"required,min=1,max=255"`
	Plan       string `json:"plan"        binding:"required,oneof=neo pro ultra"`
}

type UpdateTenantRequest struct {
	Label string `json:"label" binding:"required,min=1,max=255"`
}

type TenantResponse struct {
	ID         uuid.UUID `json:"id"`
	TenantName string    `json:"tenant_name"`
	Label      string    `json:"label"`
	Plan       string    `json:"plan"`
	Banned     bool      `json:"banned"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type SetTenantAdminRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

func ToTenantResponse(t *model.Tenant) TenantResponse {
	return TenantResponse{
		ID:         t.UUID,
		TenantName: t.TenantName,
		Label:      t.Label,
		Plan:       t.Plan,
		Banned:     t.Banned,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}
}

func ToTenantResponseList(tenants []model.Tenant) []TenantResponse {
	res := make([]TenantResponse, len(tenants))
	for i := range tenants {
		res[i] = ToTenantResponse(&tenants[i])
	}
	return res
}
