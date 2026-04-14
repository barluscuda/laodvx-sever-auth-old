package repository

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemAdminRepository struct {
	db *gorm.DB
}

func NewSystemAdminRepository(db *gorm.DB) ports.SystemAdminRepository {
	return &systemAdminRepository{db: db}
}

func (r *systemAdminRepository) Create(u *model.SystemAdmin) error {
	return r.db.Create(u).Error
}

func (r *systemAdminRepository) GetAll() ([]model.SystemAdmin, error) {
	var admins []model.SystemAdmin
	if err := r.db.Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (r *systemAdminRepository) GetByUsername(username string) (*model.SystemAdmin, error) {
	var u model.SystemAdmin
	if err := firstOrNotFound(r.db, &u, "username = ?", username); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *systemAdminRepository) GetByID(id uuid.UUID) (*model.SystemAdmin, error) {
	var u model.SystemAdmin
	if err := firstOrNotFound(r.db, &u, "uuid = ?", id); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *systemAdminRepository) Delete(id uuid.UUID) error {
	return requireAffected(r.db.Where("uuid = ?", id).Delete(&model.SystemAdmin{}))
}
