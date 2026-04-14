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

type systemAuthService struct {
	systemUserRepo ports.SystemAdminRepository
	rtRepo         ports.RefreshTokenRepository
	attemptRepo    ports.LoginAttemptRepository
	jwtCfg         config.JWTConfig
}

func NewSystemAuthService(
	systemUserRepo ports.SystemAdminRepository,
	rtRepo ports.RefreshTokenRepository,
	attemptRepo ports.LoginAttemptRepository,
	jwtCfg config.JWTConfig,
) ports.SystemAuthService {
	return &systemAuthService{systemUserRepo: systemUserRepo, rtRepo: rtRepo, attemptRepo: attemptRepo, jwtCfg: jwtCfg}
}

func (s *systemAuthService) Login(username, password string) (*ports.TokenPair, error) {
	key := fmt.Sprintf("sysadmin:%s", username)

	locked, retryAfter, err := s.attemptRepo.IsLocked(key)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, &ports.AccountLockedError{RetryAfter: retryAfter}
	}

	u, err := s.systemUserRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Run bcrypt even on miss to prevent timing-based username enumeration.
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
	return s.issuePair(u.UUID)
}

func (s *systemAuthService) Refresh(refreshToken string) (*ports.TokenPair, error) {
	claims := &token.RefreshClaims{}
	t, err := jwt.ParseWithClaims(refreshToken, claims, s.jwtCfg.RefreshKeys.KeyFunc)
	if err != nil || !t.Valid {
		return nil, ports.ErrInvalidToken
	}

	if claims.Role != ports.RoleSystemAdmin {
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

	return s.issuePair(rt.OwnerID)
}

func (s *systemAuthService) issuePair(userID uuid.UUID) (*ports.TokenPair, error) {
	return issueTokenPair(s.rtRepo, s.jwtCfg, userID, nil, ports.RoleSystemAdmin)
}
