package handler

import (
	"rag-searchbot-backend/internal/models"

	"github.com/google/uuid"
)

type UserRole string

const (
	NormalUser UserRole = "NORMAL_USER"
	WriterUser UserRole = "WRITER_USER"
	AdminUser  UserRole = "ADMIN_USER"
)

type MeResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Role      UserRole  `json:"role"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Bio       string    `json:"bio"`
	UserName  string    `json:"username"`
	NewUser   bool      `json:"new_user"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	WarpKey   string    `json:"warp_key,omitempty"` // Optional field for warp key
}

func MapResponse(user *models.User) MeResponse {
	return MeResponse{
		ID:        user.ID,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Role:      UserRole(user.Role), // Cast if handler and model enums are different
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Bio:       user.Bio,
		NewUser:   user.NewUser,
		UserName:  user.UserName,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
