package service

import (
	"time"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/pkg/token"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func issueTokenPair(
	rtRepo ports.RefreshTokenRepository,
	jwtCfg config.JWTConfig,
	userID uuid.UUID,
	tenantID *uuid.UUID,
	role string,
) (*ports.TokenPair, error) {
	now := time.Now()
	userIDStr := userID.String()
	tenantIDStr := ""
	if tenantID != nil {
		tenantIDStr = tenantID.String()
	}

	accessClaims := &token.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userIDStr,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtCfg.AccessExpiry)),
		},
		UserID:   userIDStr,
		TenantID: tenantIDStr,
		Role:     role,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodES256, accessClaims)
	at.Header["kid"] = jwtCfg.AccessKeys.ActiveKID()
	accessToken, err := at.SignedString(jwtCfg.AccessKeys.ActiveKey())
	if err != nil {
		return nil, err
	}

	tid := uuid.New()
	refreshExpiry := now.Add(jwtCfg.RefreshExpiry)
	refreshClaims := &token.RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userIDStr,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
		},
		TID:      tid.String(),
		UserID:   userIDStr,
		TenantID: tenantIDStr,
		Role:     role,
	}
	rtJWT := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims)
	rtJWT.Header["kid"] = jwtCfg.RefreshKeys.ActiveKID()
	refreshToken, err := rtJWT.SignedString(jwtCfg.RefreshKeys.ActiveKey())
	if err != nil {
		return nil, err
	}

	rt := &model.RefreshToken{
		TID:      tid,
		OwnerID:  userID,
		TenantID: tenantID,
		ExpiresAt: refreshExpiry,
	}
	if err := rtRepo.Create(rt); err != nil {
		return nil, err
	}

	return &ports.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
