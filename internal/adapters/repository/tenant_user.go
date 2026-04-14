package repository

import (
	"errors"
	"strings"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tenantUserRepository struct {
	db *gorm.DB
}

func NewTenantUserRepository(db *gorm.DB) ports.SystemUserRepository {
	return &tenantUserRepository{db: db}
}

func (r *tenantUserRepository) Create(u *model.TenantUser) error {
	if err := r.db.Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return ports.ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *tenantUserRepository) GetByID(tenantID uuid.UUID, id uuid.UUID) (*model.TenantUser, error) {
	var u model.TenantUser
	if err := r.db.First(&u, "tenant_id = ? AND uuid = ?", tenantID, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *tenantUserRepository) GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	var u model.TenantUser
	if err := r.db.First(&u, "tenant_id = ? AND email = ?", tenantID, email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *tenantUserRepository) GetAll(tenantID uuid.UUID) ([]model.TenantUser, error) {
	var users []model.TenantUser
	if err := r.db.Where("tenant_id = ?", tenantID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *tenantUserRepository) SystemGetAll() ([]model.TenantUser, error) {
	var users []model.TenantUser
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *tenantUserRepository) SystemGetAllByEmail(email string) ([]model.TenantUser, error) {
	var users []model.TenantUser
	if err := r.db.Where("email = ?", email).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *tenantUserRepository) SystemGetByID(id uuid.UUID) (*model.TenantUser, error) {
	var u model.TenantUser
	if err := r.db.First(&u, "uuid = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *tenantUserRepository) SystemSetRole(id uuid.UUID, role string) error {
	result := r.db.Model(&model.TenantUser{}).Where("uuid = ?", id).Update("role", role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrNotFound
	}
	return nil
}
