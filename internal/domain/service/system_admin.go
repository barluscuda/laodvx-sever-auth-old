package service

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
)

type systemAdminService struct {
	repo       ports.SystemUserRepository
	tenantRepo ports.TenantRepository
}

func NewSystemAdminService(repo ports.SystemUserRepository, tenantRepo ports.TenantRepository) ports.SystemAdminService {
	return &systemAdminService{repo: repo, tenantRepo: tenantRepo}
}

func (s *systemAdminService) GetAll() ([]model.TenantUser, error) {
	return s.repo.SystemGetAll()
}

func (s *systemAdminService) GetAllByTenantName(tenantName string) ([]model.TenantUser, error) {
	t, err := s.tenantRepo.GetByTenantName(tenantName)
	if err != nil {
		return nil, err
	}
	return s.repo.GetAll(t.UUID)
}

func (s *systemAdminService) GetAllByEmail(email string) ([]model.TenantUser, error) {
	return s.repo.SystemGetAllByEmail(email)
}

func (s *systemAdminService) GetByTenantAndEmail(tenantName string, email string) (*model.TenantUser, error) {
	t, err := s.tenantRepo.GetByTenantName(tenantName)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByEmail(t.UUID, email)
}

func (s *systemAdminService) GetByID(id uuid.UUID) (*model.TenantUser, error) {
	return s.repo.SystemGetByID(id)
}

func (s *systemAdminService) SetTenantAdmin(tenantName string, userID uuid.UUID) error {
	t, err := s.tenantRepo.GetByTenantName(tenantName)
	if err != nil {
		return err
	}
	// Verify the user belongs to the specified tenant before promoting.
	if _, err := s.repo.GetByID(t.UUID, userID); err != nil {
		return err
	}
	return s.repo.SystemSetRole(userID, ports.RoleTenantAdmin)
}
