package model

import "time"

type Item struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Stock     int       `json:"stock"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type ItemRequest struct {
	Code     string `json:"code"     validate:"required,max=20"`
	Name     string `json:"name"     validate:"required,max=100"`
	Category string `json:"category" validate:"required,max=50"`
	Stock    int    `json:"stock"    validate:"min=0"`
	IsActive *bool  `json:"is_active"`
}
