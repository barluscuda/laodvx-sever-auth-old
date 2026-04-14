package service

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
)

type tenantAdminService struct {
	repo ports.TenantUserRepository
}

func NewTenantAdminService(repo ports.TenantUserRepository) ports.TenantAdminService {
	return &tenantAdminService{repo: repo}
}

func (s *tenantAdminService) GetAll(tenantID uuid.UUID) ([]model.TenantUser, error) {
	return s.repo.GetAll(tenantID)
}

func (s *tenantAdminService) GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	return s.repo.GetByEmail(tenantID, email)
}

func (s *tenantAdminService) GetByID(tenantID uuid.UUID, id uuid.UUID) (*model.TenantUser, error) {
	return s.repo.GetByID(tenantID, id)
}
