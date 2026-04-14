package service

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
)

type tenantService struct {
	repo ports.TenantRepository
}

func NewTenantService(repo ports.TenantRepository) ports.TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) Create(name string) (*model.Tenant, error) {
	t := &model.Tenant{Name: name}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *tenantService) GetAll() ([]model.Tenant, error) {
	return s.repo.GetAll()
}

func (s *tenantService) GetByID(id uuid.UUID) (*model.Tenant, error) {
	return s.repo.GetByID(id)
}

func (s *tenantService) Update(id uuid.UUID, name string) (*model.Tenant, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	t.Name = name
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *tenantService) Delete(id uuid.UUID) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
