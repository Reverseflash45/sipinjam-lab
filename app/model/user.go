package model

import "time"

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	IdentityNo   string    `json:"identity_no"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Name       string `json:"name"        validate:"required,min=2,max=100"`
	Email      string `json:"email"       validate:"required,email,max=120"`
	IdentityNo string `json:"identity_no" validate:"required,min=5,max=20"`
	Password   string `json:"password"    validate:"required,min=8,max=72"`
	Role       string `json:"role"        validate:"required,oneof=dosen mahasiswa"`
}

type LoginRequest struct {
	IdentityNo string `json:"identity_no" validate:"required"`
	Password   string `json:"password"    validate:"required"`
}