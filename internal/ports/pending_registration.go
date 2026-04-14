package ports

import (
	"time"

	"github.com/google/uuid"
)

// PendingRegistration holds an unverified registration waiting for OTP confirmation.
type PendingRegistration struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	Email          string    `json:"email"`
	Password       string    `json:"password"` // bcrypt hash
	OTP            string    `json:"otp"`
	ExpiresAt      time.Time `json:"expires_at"`
	IssueCount   int       `json:"issue_count"`
	LastIssuedAt time.Time `json:"last_issued_at"`
}

// PendingRegistrationStore persists and retrieves pending (pre-verification) registrations.
type PendingRegistrationStore interface {
	Set(reg *PendingRegistration, ttl time.Duration) error
	Get(tenantID uuid.UUID, email string) (*PendingRegistration, error)
	Delete(tenantID uuid.UUID, email string) error
}
