package service

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
)

type systemAdminService struct {
	repo ports.TenantUserRepository
}

func NewSystemAdminService(repo ports.TenantUserRepository) ports.SystemAdminService {
	return &systemAdminService{repo: repo}
}

func (s *systemAdminService) GetAll() ([]model.TenantUser, error) {
	return s.repo.SystemGetAll()
}

func (s *systemAdminService) GetAllByTenantID(tenantID uuid.UUID) ([]model.TenantUser, error) {
	return s.repo.GetAll(tenantID)
}

func (s *systemAdminService) GetAllByEmail(email string) ([]model.TenantUser, error) {
	return s.repo.SystemGetAllByEmail(email)
}

func (s *systemAdminService) GetByTenantAndEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	return s.repo.GetByEmail(tenantID, email)
}

func (s *systemAdminService) GetByID(id uuid.UUID) (*model.TenantUser, error) {
	return s.repo.SystemGetByID(id)
}
