package service

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
)

type systemAdminService struct {
	systemUserRepo ports.SystemUserRepository
	tenantUserRepo ports.TenantUserRepository
	tenantRepo     ports.TenantRepository
}

func NewSystemAdminService(
	systemUserRepo ports.SystemUserRepository,
	tenantUserRepo ports.TenantUserRepository,
	tenantRepo ports.TenantRepository,
) ports.SystemAdminService {
	return &systemAdminService{
		systemUserRepo: systemUserRepo,
		tenantUserRepo: tenantUserRepo,
		tenantRepo:     tenantRepo,
	}
}

func (s *systemAdminService) GetAll() ([]model.TenantUser, error) {
	return s.systemUserRepo.SystemGetAll()
}

func (s *systemAdminService) GetAllByTenantName(tenantName string) ([]model.TenantUser, error) {
	t, err := s.tenantRepo.GetByTenantName(tenantName)
	if err != nil {
		return nil, err
	}
	return s.tenantUserRepo.GetAll(t.UUID)
}

func (s *systemAdminService) GetAllByEmail(email string) ([]model.TenantUser, error) {
	return s.systemUserRepo.SystemGetAllByEmail(email)
}

func (s *systemAdminService) GetByTenantAndEmail(tenantName string, email string) (*model.TenantUser, error) {
	t, err := s.tenantRepo.GetByTenantName(tenantName)
	if err != nil {
		return nil, err
	}
	return s.tenantUserRepo.GetByEmail(t.UUID, email)
}

func (s *systemAdminService) GetByID(id uuid.UUID) (*model.TenantUser, error) {
	return s.systemUserRepo.SystemGetByID(id)
}

func (s *systemAdminService) SetTenantAdmin(tenantName string, userID uuid.UUID) error {
	t, err := s.tenantRepo.GetByTenantName(tenantName)
	if err != nil {
		return err
	}
	if _, err := s.tenantUserRepo.GetByID(t.UUID, userID); err != nil {
		return err
	}
	return s.systemUserRepo.SystemSetRole(userID, ports.RoleTenantAdmin)
}
