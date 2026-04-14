package service

import (
	"errors"
	"fmt"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/pkg/token"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// dummyHash is used to equalize timing when a user is not found,
// preventing email enumeration via response-time analysis.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-timing-equalization"), 12)

type authService struct {
	userRepo    ports.TenantUserRepository
	rtRepo      ports.RefreshTokenRepository
	attemptRepo ports.LoginAttemptRepository
	jwtCfg      config.JWTConfig
}

func NewAuthService(
	userRepo ports.TenantUserRepository,
	rtRepo ports.RefreshTokenRepository,
	attemptRepo ports.LoginAttemptRepository,
	jwtCfg config.JWTConfig,
) ports.TenantUserAuthService {
	return &authService{userRepo: userRepo, rtRepo: rtRepo, attemptRepo: attemptRepo, jwtCfg: jwtCfg}
}

func (s *authService) Login(tenantID uuid.UUID, email, password string) (*ports.TokenPair, error) {
	key := fmt.Sprintf("user:%s:%s", tenantID, email)

	locked, retryAfter, err := s.attemptRepo.IsLocked(key)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, &ports.AccountLockedError{RetryAfter: retryAfter}
	}

	u, err := s.userRepo.GetByEmail(tenantID, email)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Run bcrypt even on miss to prevent timing-based email enumeration.
			bcrypt.CompareHashAndPassword(dummyHash, []byte(password)) //nolint:errcheck
			if err := s.attemptRepo.Record(key); err != nil {
				return nil, err
			}
			return nil, ports.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		if recordErr := s.attemptRepo.Record(key); recordErr != nil {
			return nil, recordErr
		}
		return nil, ports.ErrInvalidCredentials
	}

	if err := s.attemptRepo.Reset(key); err != nil {
		return nil, err
	}
	return s.issuePair(u.UUID, &tenantID, u.Role)
}

func (s *authService) Refresh(refreshToken string) (*ports.TokenPair, error) {
	claims := &token.RefreshClaims{}
	t, err := jwt.ParseWithClaims(refreshToken, claims, s.jwtCfg.RefreshKeys.KeyFunc)
	if err != nil || !t.Valid {
		return nil, ports.ErrInvalidToken
	}

	if claims.Role != ports.RoleUser && claims.Role != ports.RoleTenantAdmin {
		return nil, ports.ErrInvalidToken
	}

	tid, err := uuid.Parse(claims.TID)
	if err != nil {
		return nil, ports.ErrInvalidToken
	}

	rt, err := s.rtRepo.ClaimToken(tid)
	if err != nil {
		return nil, err
	}

	if rt.TenantID == nil {
		return nil, ports.ErrInvalidToken
	}

	// Re-fetch user to pick up current role
	u, err := s.userRepo.GetByID(*rt.TenantID, rt.OwnerID)
	if err != nil {
		return nil, err
	}

	return s.issuePair(rt.OwnerID, rt.TenantID, u.Role)
}

func (s *authService) issuePair(userID uuid.UUID, tenantID *uuid.UUID, role string) (*ports.TokenPair, error) {
	return issueTokenPair(s.rtRepo, s.jwtCfg, userID, tenantID, role)
}
