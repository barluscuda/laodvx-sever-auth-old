package repository

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tenantUserRepository struct {
	db *gorm.DB
}

func NewTenantUserRepository(db *gorm.DB) ports.TenantUserRepository {
	return &tenantUserRepository{db: db}
}

func (r *tenantUserRepository) Create(u *model.TenantUser) error {
	if err := r.db.Create(u).Error; err != nil {
		if isDuplicateKeyErr(err) {
			return ports.ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *tenantUserRepository) GetByID(tenantID uuid.UUID, id uuid.UUID) (*model.TenantUser, error) {
	var u model.TenantUser
	if err := firstOrNotFound(r.db, &u, "tenant_id = ? AND uuid = ?", tenantID, id); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *tenantUserRepository) GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	var u model.TenantUser
	if err := firstOrNotFound(r.db, &u, "tenant_id = ? AND email = ?", tenantID, email); err != nil {
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
