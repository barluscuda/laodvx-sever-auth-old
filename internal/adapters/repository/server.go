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

func (r *tenantRepository) GetByTenantName(tenantName string) (*model.Tenant, error) {
	var t model.Tenant
	if err := r.db.First(&t, "tenant_name = ?", tenantName).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *tenantRepository) Update(t *model.Tenant) error {
	return r.db.Model(&model.Tenant{}).Where("tenant_name = ?", t.TenantName).
		Updates(map[string]any{"label": t.Label}).Error
}

func (r *tenantRepository) Delete(tenantName string) error {
	return r.db.Where("tenant_name = ?", tenantName).Delete(&model.Tenant{}).Error
}

func (r *tenantRepository) ExistsByName(tenantName string) (bool, error) {
	var t model.Tenant
	err := r.db.Select("tenant_name").First(&t, "tenant_name = ?", tenantName).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *tenantRepository) ExistsByUUID(id uuid.UUID) (bool, error) {
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

func (r *tenantRepository) Ban(tenantName string) error {
	result := r.db.Model(&model.Tenant{}).Where("tenant_name = ?", tenantName).Update("banned", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrNotFound
	}
	return nil
}
