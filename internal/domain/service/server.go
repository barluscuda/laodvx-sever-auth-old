package service

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
)

type tenantService struct {
	repo ports.TenantRepository
}

func NewTenantService(repo ports.TenantRepository) ports.TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) Create(tenantName, label, plan string) (*model.Tenant, error) {
	t := &model.Tenant{
		TenantName: tenantName,
		Label:      label,
		Plan:       plan,
	}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *tenantService) GetAll() ([]model.Tenant, error) {
	return s.repo.GetAll()
}

func (s *tenantService) GetByTenantName(tenantName string) (*model.Tenant, error) {
	return s.repo.GetByTenantName(tenantName)
}

func (s *tenantService) Update(tenantName, label string) (*model.Tenant, error) {
	t, err := s.repo.GetByTenantName(tenantName)
	if err != nil {
		return nil, err
	}
	t.Label = label
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *tenantService) Delete(tenantName string) error {
	if _, err := s.repo.GetByTenantName(tenantName); err != nil {
		return err
	}
	return s.repo.Delete(tenantName)
}

func (s *tenantService) Ban(tenantName string) error {
	return s.repo.Ban(tenantName)
}
