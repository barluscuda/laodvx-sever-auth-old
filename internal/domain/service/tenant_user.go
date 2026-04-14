package service

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type tenantUserService struct {
	repo            ports.TenantUserRepository
	pendingStore    ports.PendingRegistrationStore
	registerAttempt ports.RegistrationAttemptRepository
	emailSender     ports.EmailSender
	emailCfg        config.EmailConfig
}

func NewTenantUserService(
	repo ports.TenantUserRepository,
	pendingStore ports.PendingRegistrationStore,
	registerAttempt ports.RegistrationAttemptRepository,
	emailSender ports.EmailSender,
	emailCfg config.EmailConfig,
) ports.TenantUserService {
	return &tenantUserService{
		repo:            repo,
		pendingStore:    pendingStore,
		registerAttempt: registerAttempt,
		emailSender:     emailSender,
		emailCfg:        emailCfg,
	}
}

// Create hashes the password, stores a pending registration in cache, and
// sends an OTP email. The user is NOT written to the database until VerifyEmail
// succeeds.
func (s *tenantUserService) Create(tenantID uuid.UUID, email, password string) error {
	// Honour an active verify-lockout on this email — a registration is
	// locked for OTPVerifyLockoutTTL after OTPMaxVerifyAttempts bad OTPs.
	if locked, retryAfter, err := s.registerAttempt.IsLocked(registerLockKey(tenantID, email)); err != nil {
		return err
	} else if locked {
		return &ports.AccountLockedError{RetryAfter: retryAfter}
	}

	// Existence checks first — bcrypt is expensive, so never run it on a
	// request that we are going to reject anyway (DoS hardening) and so the
	// response time does not leak whether the email is registered.
	if _, err := s.repo.GetByEmail(tenantID, email); err == nil {
		return ports.ErrDuplicateEmail
	} else if !errors.Is(err, ports.ErrNotFound) {
		return err
	}

	// A pending registration means the user signed up but has not yet
	// verified their email. Surface a distinct error so the UI can prompt
	// them to check their inbox or resend the OTP instead of showing a
	// generic "already registered" message.
	if _, err := s.pendingStore.Get(tenantID, email); err == nil {
		return ports.ErrRegistrationPending
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}

	now := time.Now()
	reg := &ports.PendingRegistration{
		TenantID:     tenantID,
		Email:        email,
		Password:     string(hash),
		OTP:          otp,
		ExpiresAt:    now.Add(s.emailCfg.OTPExpiry),
		IssueCount:   1,
		LastIssuedAt: now,
	}
	if err := s.pendingStore.Set(reg, s.emailCfg.OTPExpiry); err != nil {
		return err
	}

	if err := s.emailSender.SendVerificationEmail(email, otp); err != nil {
		// Roll back the pending reg so the user can retry without hitting
		// ErrDuplicateEmail on the pending-store check.
		_ = s.pendingStore.Delete(tenantID, email)
		return err
	}
	return nil
}

// VerifyEmail validates the OTP against the pending registration, creates the
// user in the database, and removes the pending entry from cache. After
// OTPMaxVerifyAttempts wrong guesses the email is locked for
// OTPVerifyLockoutTTL via the registration-attempt store.
func (s *tenantUserService) VerifyEmail(tenantID uuid.UUID, email, otp string) error {
	lockKey := registerLockKey(tenantID, email)

	if locked, retryAfter, err := s.registerAttempt.IsLocked(lockKey); err != nil {
		return err
	} else if locked {
		return &ports.AccountLockedError{RetryAfter: retryAfter}
	}

	reg, err := s.pendingStore.Get(tenantID, email)
	if err != nil {
		// Count misses against unknown/expired pendings too so an attacker
		// cannot probe indefinitely once the pending entry TTL elapses.
		_ = s.registerAttempt.Record(lockKey)
		return ports.ErrVerificationTokenInvalid
	}

	if time.Now().After(reg.ExpiresAt) {
		s.pendingStore.Delete(tenantID, email) //nolint:errcheck
		_ = s.registerAttempt.Record(lockKey)
		return ports.ErrVerificationTokenInvalid
	}

	if subtle.ConstantTimeCompare([]byte(reg.OTP), []byte(otp)) != 1 {
		if err := s.registerAttempt.Record(lockKey); err != nil {
			return err
		}
		return ports.ErrVerificationTokenInvalid
	}

	u := &model.TenantUser{
		TenantID: tenantID,
		Email:    reg.Email,
		Password: reg.Password,
	}
	if err := s.repo.Create(u); err != nil {
		return err
	}

	s.pendingStore.Delete(tenantID, email)  //nolint:errcheck
	_ = s.registerAttempt.Reset(lockKey)
	return nil
}

// ResendVerification generates a new OTP for an existing pending registration
// and resends the verification email. Returns nil for unknown emails to avoid
// enumerating registered accounts, but enforces per-email OTP rate limits
// (cooldown + max issues) to prevent mailbox flooding.
func (s *tenantUserService) ResendVerification(tenantID uuid.UUID, email string) error {
	if locked, retryAfter, err := s.registerAttempt.IsLocked(registerLockKey(tenantID, email)); err != nil {
		return err
	} else if locked {
		return &ports.AccountLockedError{RetryAfter: retryAfter}
	}

	reg, err := s.pendingStore.Get(tenantID, email)
	if err != nil {
		// Unknown email or already verified — silent no-op.
		return nil
	}

	now := time.Now()
	if s.emailCfg.OTPResendCooldown > 0 &&
		now.Sub(reg.LastIssuedAt) < s.emailCfg.OTPResendCooldown {
		return ports.ErrOTPRateLimited
	}
	if s.emailCfg.OTPMaxIssues > 0 && reg.IssueCount >= s.emailCfg.OTPMaxIssues {
		return ports.ErrOTPRateLimited
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}
	reg.OTP = otp
	reg.ExpiresAt = now.Add(s.emailCfg.OTPExpiry)
	reg.LastIssuedAt = now
	reg.IssueCount++
	if err := s.pendingStore.Set(reg, s.emailCfg.OTPExpiry); err != nil {
		return err
	}
	return s.emailSender.SendVerificationEmail(email, otp)
}

func (s *tenantUserService) GetByID(tenantID uuid.UUID, id uuid.UUID) (*model.TenantUser, error) {
	return s.repo.GetByID(tenantID, id)
}

func (s *tenantUserService) GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error) {
	return s.repo.GetByEmail(tenantID, email)
}

// registerLockKey is the per-(tenant,email) key used by the registration
// verify-lockout store. Kept in sync with pending-registration keying so
// the same normalised email identifies both records.
func registerLockKey(tenantID uuid.UUID, email string) string {
	return fmt.Sprintf("%s:%s", tenantID, strings.ToLower(strings.TrimSpace(email)))
}

// generateOTP returns a cryptographically random 6-digit numeric string.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
