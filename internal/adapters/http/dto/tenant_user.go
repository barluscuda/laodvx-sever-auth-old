package dto

import (
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
)

type CreateTenantUserRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp"   binding:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type TenantUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToTenantUserResponse(u *model.TenantUser) TenantUserResponse {
	return TenantUserResponse{
		ID:        u.UUID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func ToTenantUserResponseList(users []model.TenantUser) []TenantUserResponse {
	res := make([]TenantUserResponse, len(users))
	for i := range users {
		res[i] = ToTenantUserResponse(&users[i])
	}
	return res
}
