package repository

import (
	"errors"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) ports.TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(t *model.Tenant) error {
	return r.db.Create(t).Error
}

func (r *tenantRepository) GetAll() ([]model.Tenant, error) {
	var tenants []model.Tenant
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *tenantRepository) GetByID(id uuid.UUID) (*model.Tenant, error) {
	var t model.Tenant
	if err := r.db.First(&t, "uuid = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *tenantRepository) GetByName(name string) (*model.Tenant, error) {
	var t model.Tenant
	if err := r.db.First(&t, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *tenantRepository) Update(t *model.Tenant) error {
	return r.db.Model(&model.Tenant{}).Where("uuid = ?", t.UUID).Update("name", t.Name).Error
}

func (r *tenantRepository) Delete(id uuid.UUID) error {
	return r.db.Where("uuid = ?", id).Delete(&model.Tenant{}).Error
}

func (r *tenantRepository) ExistsByID(id uuid.UUID) (bool, error) {
	var t model.Tenant
	err := r.db.Select("uuid").First(&t, "uuid = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
