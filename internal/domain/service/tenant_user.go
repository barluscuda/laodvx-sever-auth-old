package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type tenantUserService struct {
	repo         ports.TenantUserRepository
	pendingStore ports.PendingRegistrationStore
	emailSender  ports.EmailSender
	emailCfg     config.EmailConfig
}

func NewTenantUserService(
	repo ports.TenantUserRepository,
	pendingStore ports.PendingRegistrationStore,
	emailSender ports.EmailSender,
	emailCfg config.EmailConfig,
) ports.TenantUserService {
	return &tenantUserService{
		repo:         repo,
		pendingStore: pendingStore,
		emailSender:  emailSender,
		emailCfg:     emailCfg,
	}
}

// Create hashes the password, stores a pending registration in cache, and
// sends an OTP email. The user is NOT written to the database until VerifyEmail
// succeeds.
func (s *tenantUserService) Create(tenantID uuid.UUID, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// Reject if a verified user already exists in the database.
	if _, err := s.repo.GetByEmail(tenantID, email); err == nil {
		return ports.ErrDuplicateEmail
	} else if !errors.Is(err, ports.ErrNotFound) {
		return err
	}

	// Reject if a pending (unverified) registration already exists in cache.
	if _, err := s.pendingStore.Get(tenantID, email); err == nil {
		return ports.ErrDuplicateEmail
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}

	reg := &ports.PendingRegistration{
		TenantID:  tenantID,
		Email:     email,
		Password:  string(hash),
		OTP:       otp,
		ExpiresAt: time.Now().Add(s.emailCfg.OTPExpiry),
	}
	if err := s.pendingStore.Set(reg, s.emailCfg.OTPExpiry); err != nil {
		return err
	}

	return s.emailSender.SendVerificationEmail(email, otp)
}

// VerifyEmail validates the OTP against the pending registration, creates the
// user in the database, and removes the pending entry from cache.
func (s *tenantUserService) VerifyEmail(tenantID uuid.UUID, email, otp string) error {
	reg, err := s.pendingStore.Get(tenantID, email)
	if err != nil {
		return ports.ErrVerificationTokenInvalid
	}

	if time.Now().After(reg.ExpiresAt) {
		s.pendingStore.Delete(tenantID, email) //nolint:errcheck
		return ports.ErrVerificationTokenInvalid
	}

	if reg.OTP != otp {
		return ports.ErrVerificationTokenInvalid
	}

	u := &model.TenantUser{
		TenantID:      tenantID,
		Email:         reg.Email,
		Password:      reg.Password,
		EmailVerified: true,
	}
	if err := s.repo.Create(u); err != nil {
		return err
	}

	s.pendingStore.Delete(tenantID, email) //nolint:errcheck
	return nil
}

// ResendVerification generates a new OTP for an existing pending registration
// and resends the verification email. Always returns nil to avoid enumerating
// registered emails.
func (s *tenantUserService) ResendVerification(tenantID uuid.UUID, email string) error {
	reg, err := s.pendingStore.Get(tenantID, email)
	if err != nil {
		// Unknown email or already verified — silent no-op.
		return nil
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}
	reg.OTP = otp
	reg.ExpiresAt = time.Now().Add(s.emailCfg.OTPExpiry)
	if err := s.pendingStore.Set(reg, s.emailCfg.OTPExpiry); err != nil {
		return err
	}
	return s.emailSender.SendVerificationEmail(email, otp)
}

func (s *tenantUserService) GetByID(tenantID, id uuid.UUID) (*model.TenantUser, error) {
	return s.repo.GetByID(tenantID, id)
}

func (s *tenantUserService) GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	return s.repo.GetByEmail(tenantID, email)
}

func (s *tenantUserService) GetAllServer() ([]model.TenantUser, error) {
	return s.repo.SystemGetAll()
}

// generateOTP returns a cryptographically random 6-digit numeric string.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
