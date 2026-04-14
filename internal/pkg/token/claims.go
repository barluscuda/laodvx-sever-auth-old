package token

import "github.com/golang-jwt/jwt/v5"

// Claims is embedded in access tokens.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"uid"`
	TenantID string `json:"ten,omitempty"` // empty for system admin
	Role     string `json:"role"`
}

// RefreshClaims is embedded in refresh tokens.
// TID (token ID) is stored in the database for one-time-use enforcement.
type RefreshClaims struct {
	jwt.RegisteredClaims
	TID      string `json:"tid"`
	UserID   string `json:"uid"`
	TenantID string `json:"ten,omitempty"`
	Role     string `json:"role"`
}
