package repository

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemGlobalUserRepository struct {
	db *gorm.DB
}

func NewSystemGlobalUserRepository(db *gorm.DB) ports.SystemUserRepository {
	return &systemGlobalUserRepository{db: db}
}

func (r *systemGlobalUserRepository) SystemGetAll() ([]model.TenantUser, error) {
	var users []model.TenantUser
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *systemGlobalUserRepository) SystemGetAllByEmail(email string) ([]model.TenantUser, error) {
	var users []model.TenantUser
	if err := r.db.Where("email = ?", email).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *systemGlobalUserRepository) SystemGetByID(id uuid.UUID) (*model.TenantUser, error) {
	var u model.TenantUser
	if err := firstOrNotFound(r.db, &u, "uuid = ?", id); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *systemGlobalUserRepository) SystemSetRole(id uuid.UUID, role string) error {
	return requireAffected(r.db.Model(&model.TenantUser{}).
		Where("uuid = ?", id).
		Update("role", role))
}
