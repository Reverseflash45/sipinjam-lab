package model

import "time"

// BookingStatus adalah status yang valid untuk sebuah peminjaman.
type BookingStatus string

const (
	StatusPending   BookingStatus = "pending"
	StatusApproved  BookingStatus = "approved"
	StatusRejected  BookingStatus = "rejected"
	StatusCancelled BookingStatus = "cancelled"
	StatusReturned  BookingStatus = "returned"
)

// Booking adalah satu pengajuan peminjaman. Boleh berisi ruang saja,
// alat saja, atau keduanya.
type Booking struct {
	ID         int           `json:"id"`
	UserID     int           `json:"user_id"`
	RoomID     *int          `json:"room_id,omitempty"`
	Purpose    string        `json:"purpose"`
	StartAt    time.Time     `json:"start_at"`
	EndAt      time.Time     `json:"end_at"`
	Status     BookingStatus `json:"status"`
	ReviewedBy *int          `json:"reviewed_by,omitempty"`
	ReviewNote *string       `json:"review_note,omitempty"`
	ReturnedAt *time.Time    `json:"returned_at,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`

	Items []BookingItem `json:"items,omitempty"`
}

// BookingItem adalah satu baris alat di dalam pengajuan.
type BookingItem struct {
	ItemID   int `json:"item_id"  validate:"required,gt=0"`
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

// CreateBookingRequest adalah body untuk POST /api/v1/bookings.
// StartAt/EndAt harus dalam format RFC3339, contoh: 2026-12-01T10:00:00+07:00
type CreateBookingRequest struct {
	RoomID  *int          `json:"room_id"  validate:"omitempty,gt=0"`
	Purpose string        `json:"purpose"  validate:"required,min=3,max=200"`
	StartAt time.Time     `json:"start_at" validate:"required"`
	EndAt   time.Time     `json:"end_at"   validate:"required,gtfield=StartAt"`
	Items   []BookingItem `json:"items"    validate:"omitempty,dive"`
}