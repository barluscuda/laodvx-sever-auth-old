package service

import (
	"errors"
	"fmt"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/pkg/token"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type tenantUserAuthService struct {
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
	return &tenantUserAuthService{userRepo: userRepo, rtRepo: rtRepo, attemptRepo: attemptRepo, jwtCfg: jwtCfg}
}

func (s *tenantUserAuthService) Login(tenantID uuid.UUID, email, password string) (*ports.TokenPair, error) {
	key := fmt.Sprintf("user:%s:%s", tenantID, email)

	var user = struct {
		id   uuid.UUID
		role string
	}{}

	err := verifyCredentials(s.attemptRepo, key, password, func() (string, error) {
		u, err := s.userRepo.GetByEmail(tenantID, email)
		if err != nil {
			return "", err
		}
		user.id, user.role = u.UUID, u.Role
		return u.Password, nil
	})
	if err != nil {
		return nil, err
	}

	return issueTokenPair(s.rtRepo, s.jwtCfg, user.id, &tenantID, user.role)
}

func (s *tenantUserAuthService) Refresh(refreshToken string) (*ports.TokenPair, error) {
	claims, err := parseRefreshClaims(refreshToken, s.jwtCfg, ports.RoleUser, ports.RoleTenantAdmin)
	if err != nil {
		return nil, err
	}

	tid, err := uuid.Parse(claims.TID)
	if err != nil {
		return nil, ports.ErrInvalidToken
	}

	rt, err := s.rtRepo.ClaimToken(tid)
	if err != nil {
		return nil, normalizeRefreshErr(err)
	}
	if rt.TenantID == nil {
		return nil, ports.ErrInvalidToken
	}

	// Re-fetch user to pick up the current role.
	u, err := s.userRepo.GetByID(*rt.TenantID, rt.OwnerID)
	if err != nil {
		return nil, normalizeRefreshErr(err)
	}
	return issueTokenPair(s.rtRepo, s.jwtCfg, rt.OwnerID, rt.TenantID, u.Role)
}

// normalizeRefreshErr collapses ErrNotFound into ErrInvalidToken so callers
// don't leak whether a refresh token or its owning user actually exists.
func normalizeRefreshErr(err error) error {
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ErrInvalidToken
	}
	return err
}

// parseRefreshClaims validates the refresh-token JWT and checks its role is in the allowed set.
func parseRefreshClaims(refreshToken string, jwtCfg config.JWTConfig, allowedRoles ...string) (*token.RefreshClaims, error) {
	claims := &token.RefreshClaims{}
	t, err := jwt.ParseWithClaims(refreshToken, claims, jwtCfg.RefreshKeys.KeyFunc)
	if err != nil || !t.Valid {
		return nil, ports.ErrInvalidToken
	}
	for _, r := range allowedRoles {
		if claims.Role == r {
			return claims, nil
		}
	}
	return nil, ports.ErrInvalidToken
}
