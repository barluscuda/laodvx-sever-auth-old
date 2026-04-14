package service

import (
	"fmt"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/google/uuid"
)

type systemAuthService struct {
	adminRepo   ports.SystemAdminRepository
	rtRepo      ports.RefreshTokenRepository
	attemptRepo ports.LoginAttemptRepository
	jwtCfg      config.JWTConfig
}

func NewSystemAuthService(
	adminRepo ports.SystemAdminRepository,
	rtRepo ports.RefreshTokenRepository,
	attemptRepo ports.LoginAttemptRepository,
	jwtCfg config.JWTConfig,
) ports.SystemAuthService {
	return &systemAuthService{adminRepo: adminRepo, rtRepo: rtRepo, attemptRepo: attemptRepo, jwtCfg: jwtCfg}
}

func (s *systemAuthService) Login(username, password string) (*ports.TokenPair, error) {
	key := fmt.Sprintf("sysadmin:%s", username)

	var adminID uuid.UUID
	err := verifyCredentials(s.attemptRepo, key, password, func() (string, error) {
		u, err := s.adminRepo.GetByUsername(username)
		if err != nil {
			return "", err
		}
		adminID = u.UUID
		return u.Password, nil
	})
	if err != nil {
		return nil, err
	}

	return issueTokenPair(s.rtRepo, s.jwtCfg, adminID, nil, ports.RoleSystemAdmin)
}

func (s *systemAuthService) Refresh(refreshToken string) (*ports.TokenPair, error) {
	claims, err := parseRefreshClaims(refreshToken, s.jwtCfg, ports.RoleSystemAdmin)
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
	return issueTokenPair(s.rtRepo, s.jwtCfg, rt.OwnerID, nil, ports.RoleSystemAdmin)
}
