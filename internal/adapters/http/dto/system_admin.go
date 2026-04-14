package dto

import (
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/google/uuid"
)

type CreateSystemAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=8"`
}

type SystemAdminResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToSystemAdminResponse(a *model.SystemAdmin) SystemAdminResponse {
	return SystemAdminResponse{
		ID:        a.UUID,
		Username:  a.Username,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func ToSystemAdminResponseList(admins []model.SystemAdmin) []SystemAdminResponse {
	res := make([]SystemAdminResponse, len(admins))
	for i := range admins {
		res[i] = ToSystemAdminResponse(&admins[i])
	}
	return res
}
