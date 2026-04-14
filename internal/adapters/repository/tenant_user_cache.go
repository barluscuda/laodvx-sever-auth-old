package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// cachedUser is the on-the-wire representation of a TenantUser in Redis.
// Password is optional so ID-only cache entries omit it.
type cachedUser struct {
	UUID      uuid.UUID  `json:"uuid"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Email     string     `json:"email"`
	Password  string     `json:"password,omitempty"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// cachedUserEntry lets us cache negative lookups ("not found") so repeated
// misses don't hammer the database.
type cachedUserEntry struct {
	Found bool        `json:"found"`
	User  *cachedUser `json:"user,omitempty"`
}

type cachedTenantUserRepository struct {
	repo ports.SystemUserRepository
	rdb  *redis.Client
	ttl  time.Duration
}

func NewCachedTenantUserRepository(repo ports.SystemUserRepository, rdb *redis.Client, ttl time.Duration) ports.SystemUserRepository {
	return &cachedTenantUserRepository{repo: repo, rdb: rdb, ttl: ttl}
}

func (r *cachedTenantUserRepository) Create(u *model.TenantUser) error {
	if err := r.repo.Create(u); err != nil {
		return err
	}
	r.cacheUser(u)
	return nil
}

func (r *cachedTenantUserRepository) GetByID(tenantID uuid.UUID, id uuid.UUID) (*model.TenantUser, error) {
	return r.fetchAndCache(
		userIDKey(tenantID, id),
		false, // no password needed for ID lookups
		func() (*model.TenantUser, error) { return r.repo.GetByID(tenantID, id) },
	)
}

func (r *cachedTenantUserRepository) GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	return r.fetchAndCache(
		userEmailKey(tenantID, email),
		true, // email lookups are used for login, so keep password
		func() (*model.TenantUser, error) { return r.repo.GetByEmail(tenantID, email) },
	)
}

func (r *cachedTenantUserRepository) GetAll(tenantID uuid.UUID) ([]model.TenantUser, error) {
	return r.repo.GetAll(tenantID)
}

func (r *cachedTenantUserRepository) SystemGetAll() ([]model.TenantUser, error) {
	return r.repo.SystemGetAll()
}

func (r *cachedTenantUserRepository) SystemGetAllByEmail(email string) ([]model.TenantUser, error) {
	return r.repo.SystemGetAllByEmail(email)
}

func (r *cachedTenantUserRepository) SystemGetByID(id uuid.UUID) (*model.TenantUser, error) {
	return r.fetchAndCache(
		userIDKey(uuid.Nil, id),
		false,
		func() (*model.TenantUser, error) { return r.repo.SystemGetByID(id) },
	)
}

func (r *cachedTenantUserRepository) SystemSetRole(id uuid.UUID, role string) error {
	return r.repo.SystemSetRole(id, role)
}

// fetchAndCache looks up the cache first, then falls back to the underlying
// repo and populates the cache (including negative entries for not-found).
func (r *cachedTenantUserRepository) fetchAndCache(
	key string,
	withPassword bool,
	fromRepo func() (*model.TenantUser, error),
) (*model.TenantUser, error) {
	if u, found, hit := r.readEntry(key, withPassword); hit {
		if !found {
			return nil, ports.ErrNotFound
		}
		return u, nil
	}

	u, err := fromRepo()
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			r.writeEntry(key, cachedUserEntry{Found: false})
		}
		return nil, err
	}
	r.cacheUser(u)
	return u, nil
}

// cacheUser writes both the ID-keyed and email-keyed entries for the user.
func (r *cachedTenantUserRepository) cacheUser(u *model.TenantUser) {
	r.writeEntry(userIDKey(u.TenantID, u.UUID), cachedUserEntry{Found: true, User: toCached(u, false)})
	r.writeEntry(userEmailKey(u.TenantID, u.Email), cachedUserEntry{Found: true, User: toCached(u, true)})
}

func (r *cachedTenantUserRepository) writeEntry(key string, entry cachedUserEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("user cache: marshal error for %s: %v", key, err)
		return
	}
	if err := r.rdb.Set(context.Background(), key, data, r.ttl).Err(); err != nil {
		log.Printf("user cache: set error for %s: %v", key, err)
	}
}

// readEntry returns (user, found, hit). hit=false means cache miss, caller
// should query the underlying repo.
func (r *cachedTenantUserRepository) readEntry(key string, withPassword bool) (*model.TenantUser, bool, bool) {
	data, err := r.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, false, false
	}
	var entry cachedUserEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, false
	}
	if !entry.Found {
		return nil, false, true
	}
	if entry.User == nil {
		return nil, false, false
	}
	return fromCached(entry.User, withPassword), true, true
}

func toCached(u *model.TenantUser, withPassword bool) *cachedUser {
	cu := &cachedUser{
		UUID:      u.UUID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if withPassword {
		cu.Password = u.Password
	}
	return cu
}

func fromCached(cu *cachedUser, withPassword bool) *model.TenantUser {
	u := &model.TenantUser{
		UUID:      cu.UUID,
		TenantID:  cu.TenantID,
		Email:     cu.Email,
		Role:      cu.Role,
		CreatedAt: cu.CreatedAt,
		UpdatedAt: cu.UpdatedAt,
	}
	if withPassword {
		u.Password = cu.Password
	}
	return u
}

func userIDKey(tenantID uuid.UUID, id uuid.UUID) string {
	return fmt.Sprintf("user:%s:%s", tenantID, id)
}

func userEmailKey(tenantID uuid.UUID, email string) string {
	return fmt.Sprintf("user-email:%s:%s", tenantID, strings.ToLower(strings.TrimSpace(email)))
}
