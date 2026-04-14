package service

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type devSystemAdminService struct {
	repo ports.SystemAdminRepository
}

func NewDevSystemAdminService(repo ports.SystemAdminRepository) ports.DevSystemAdminService {
	return &devSystemAdminService{repo: repo}
}

func (s *devSystemAdminService) Create(username, password string) (*model.SystemAdmin, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}
	admin := &model.SystemAdmin{
		Username: username,
		Password: string(hash),
	}
	if err := s.repo.Create(admin); err != nil {
		return nil, err
	}
	return admin, nil
}

func (s *devSystemAdminService) GetAll() ([]model.SystemAdmin, error) {
	return s.repo.GetAll()
}

func (s *devSystemAdminService) GetByID(id uuid.UUID) (*model.SystemAdmin, error) {
	return s.repo.GetByID(id)
}

func (s *devSystemAdminService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
