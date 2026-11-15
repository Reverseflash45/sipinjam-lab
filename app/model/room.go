package model

import "time"

type Room struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Building  string    `json:"building"`
	Capacity  int       `json:"capacity"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type RoomRequest struct {
	Code     string `json:"code"     validate:"required,max=20"`
	Name     string `json:"name"     validate:"required,max=100"`
	Building string `json:"building" validate:"required,max=50"`
	Capacity int    `json:"capacity" validate:"required,min=1"`
	IsActive *bool  `json:"is_active"`
}
