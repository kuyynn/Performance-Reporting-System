package model

import "time"

type User struct {
	ID           int64
	Username     string
	Email        string    `gorm:"size:80;unique;not null" json:"email"`
	PasswordHash string
	FullName     string
	RoleID       string    `gorm:"type:uuid;not null" json:"role_id"`
	Role         string
	Permissions  []string
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LogoutRequest struct {
	ID string `json:"id"`
}

type UserCreateRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	FullName string `json:"full_name" validate:"required"`
}

type UserUpdateRequest struct {
	ID       int64  `json:"id"`
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}
