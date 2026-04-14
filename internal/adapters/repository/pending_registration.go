package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type pendingRegistrationStore struct {
	rdb *redis.Client
}

func NewPendingRegistrationStore(rdb *redis.Client) ports.PendingRegistrationStore {
	return &pendingRegistrationStore{rdb: rdb}
}

func (s *pendingRegistrationStore) Set(reg *ports.PendingRegistration, ttl time.Duration) error {
	data, err := json.Marshal(reg)
	if err != nil {
		return err
	}
	return s.rdb.Set(context.Background(), pendingRegKey(reg.TenantID, reg.Email), data, ttl).Err()
}

func (s *pendingRegistrationStore) Get(tenantID uuid.UUID, email string) (*ports.PendingRegistration, error) {
	data, err := s.rdb.Get(context.Background(), pendingRegKey(tenantID, email)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	var reg ports.PendingRegistration
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

func (s *pendingRegistrationStore) Delete(tenantID uuid.UUID, email string) error {
	return s.rdb.Del(context.Background(), pendingRegKey(tenantID, email)).Err()
}

func pendingRegKey(tenantID uuid.UUID, email string) string {
	return fmt.Sprintf("pending-reg:%s:%s", tenantID, strings.ToLower(strings.TrimSpace(email)))
}
