package service

import (
	"time"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
)

const (
	MaxBookingDuration = 8 * time.Hour
	MaxDaysAhead       = 30
)

func ValidateWindow(start, end, now time.Time) map[string]string {
	errs := make(map[string]string)

	if !end.After(start) {
		errs["end_at"] = "waktu selesai harus setelah waktu mulai"
	}

	if end.Sub(start) > MaxBookingDuration {
		errs["end_at"] = "durasi peminjaman maksimal 8 jam"
	}

	if start.Before(now) {
		errs["start_at"] = "waktu mulai tidak boleh di masa lalu"
	}

	if start.After(now.AddDate(0, 0, MaxDaysAhead)) {
		errs["start_at"] = "peminjaman maksimal 30 hari ke depan"
	}

	// Jika tidak ada kesalahan, kembalikan nil.
	if len(errs) == 0 {
		return nil
	}

	return errs
}

func MaxActiveBookings(role model.Role) int {
	switch role {
	case model.RoleMahasiswa:
		return 3
	case model.RoleDosen:
		return 5
	default:
		return 0
	}
}

var transitions = map[model.BookingStatus][]model.BookingStatus{
	model.StatusPending: {
		model.StatusApproved,
		model.StatusRejected,
		model.StatusCancelled,
	},
	model.StatusApproved: {
		model.StatusReturned,
	},
}

func CanTransition(from, to model.BookingStatus) bool {
	for _, allowed := range transitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

func CanViewBooking(u model.AuthUser, ownerID int) bool {
	return u.UserID == ownerID ||
		u.Role == model.RoleLaboran ||
		u.Role == model.RoleAdmin
}
