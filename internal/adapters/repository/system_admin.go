package repository

import (
	"errors"

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
	if err := r.db.First(&u, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *systemAdminRepository) GetByID(id uuid.UUID) (*model.SystemAdmin, error) {
	var u model.SystemAdmin
	if err := r.db.First(&u, "uuid = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *systemAdminRepository) Delete(id uuid.UUID) error {
	result := r.db.Where("uuid = ?", id).Delete(&model.SystemAdmin{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrNotFound
	}
	return nil
}
