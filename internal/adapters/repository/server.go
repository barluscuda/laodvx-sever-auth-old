package repository

import (
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
	if err := firstOrNotFound(r.db, &t, "tenant_name = ?", tenantName); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tenantRepository) Update(t *model.Tenant) error {
	return r.db.Model(&model.Tenant{}).
		Where("tenant_name = ?", t.TenantName).
		Updates(map[string]any{"label": t.Label}).Error
}

func (r *tenantRepository) Delete(tenantName string) error {
	return r.db.Where("tenant_name = ?", tenantName).Delete(&model.Tenant{}).Error
}

func (r *tenantRepository) ExistsByName(tenantName string) (bool, error) {
	return existsBy(r.db, &model.Tenant{}, "tenant_name = ?", tenantName)
}

func (r *tenantRepository) ExistsByUUID(id uuid.UUID) (bool, error) {
	return existsBy(r.db, &model.Tenant{}, "uuid = ?", id)
}

func (r *tenantRepository) Ban(tenantName string) error {
	return requireAffected(r.db.Model(&model.Tenant{}).
		Where("tenant_name = ?", tenantName).
		Update("banned", true))
}
