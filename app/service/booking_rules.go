package service

import (
	"time"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
)

// Aturan peminjaman yang murni (tanpa database), supaya mudah diuji
// dengan unit test.

const (
	MaxBookingDuration = 8 * time.Hour
	MaxDaysAhead       = 30
)

// ValidateWindow memeriksa rentang waktu peminjaman. Hasil nil berarti valid.
func ValidateWindow(start, end, now time.Time) map[string]string {
	errs := map[string]string{}
	if !start.After(now) {
		errs["start_at"] = "harus di masa depan"
	}
	if start.After(now.AddDate(0, 0, MaxDaysAhead)) {
		errs["start_at"] = "paling jauh 30 hari dari sekarang"
	}
	if !end.After(start) {
		errs["end_at"] = "harus setelah start_at"
	} else if end.Sub(start) > MaxBookingDuration {
		errs["end_at"] = "lama peminjaman maksimal 8 jam"
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// MaxActiveBookings adalah batas peminjaman pending + approved per orang.
func MaxActiveBookings(role string) int {
	switch role {
	case model.RoleMahasiswa:
		return 3
	case model.RoleDosen:
		return 5
	default:
		return 0 // laboran dan admin mengelola, tidak meminjam
	}
}

// transitions mendaftar perpindahan status yang sah.
var transitions = map[string][]string{
	model.StatusPending:  {model.StatusApproved, model.StatusRejected, model.StatusCancelled},
	model.StatusApproved: {model.StatusReturned},
}

func CanTransition(from, to string) bool {
	for _, next := range transitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// CanViewBooking: pemilik selalu boleh, laboran dan admin boleh melihat semua.
func CanViewBooking(u model.AuthUser, ownerID int) bool {
	return u.UserID == ownerID || u.Role == model.RoleLaboran || u.Role == model.RoleAdmin
}
