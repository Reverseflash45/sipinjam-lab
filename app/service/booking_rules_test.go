package service

import (
	"testing"
	"time"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
)

func TestValidateWindow(t *testing.T) {
	now := time.Date(2026, 11, 16, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		start, end time.Time
		wantErrIn  string
	}{
		{"valid 2 jam besok", now.Add(24 * time.Hour), now.Add(26 * time.Hour), ""},
		{"tepat 8 jam masih boleh", now.Add(time.Hour), now.Add(9 * time.Hour), ""},
		{"lebih dari 8 jam", now.Add(time.Hour), now.Add(9*time.Hour + time.Minute), "end_at"},
		{"mulai di masa lalu", now.Add(-time.Hour), now.Add(time.Hour), "start_at"},
		{"selesai sebelum mulai", now.Add(3 * time.Hour), now.Add(2 * time.Hour), "end_at"},
		{"lebih dari 30 hari", now.AddDate(0, 0, 31), now.AddDate(0, 0, 31).Add(time.Hour), "start_at"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateWindow(tt.start, tt.end, now)
			if tt.wantErrIn == "" && errs != nil {
				t.Fatalf("seharusnya valid, dapat %v", errs)
			}
			if tt.wantErrIn != "" {
				if _, ok := errs[tt.wantErrIn]; !ok {
					t.Fatalf("seharusnya ada error di %s, dapat %v", tt.wantErrIn, errs)
				}
			}
		})
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from, to string
		want     bool
	}{
		{model.StatusPending, model.StatusApproved, true},
		{model.StatusPending, model.StatusCancelled, true},
		{model.StatusApproved, model.StatusReturned, true},
		{model.StatusApproved, model.StatusCancelled, false},
		{model.StatusRejected, model.StatusApproved, false},
		{model.StatusReturned, model.StatusApproved, false},
	}
	for _, tt := range tests {
		if got := CanTransition(tt.from, tt.to); got != tt.want {
			t.Errorf("CanTransition(%s, %s) = %v, seharusnya %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestMaxActiveBookings(t *testing.T) {
	if MaxActiveBookings(model.RoleMahasiswa) != 3 || MaxActiveBookings(model.RoleDosen) != 5 ||
		MaxActiveBookings(model.RoleLaboran) != 0 {
		t.Error("batas peminjaman aktif per role tidak sesuai rancangan")
	}
}

func TestCanViewBooking(t *testing.T) {
	owner := model.AuthUser{UserID: 7, Role: model.RoleMahasiswa}
	other := model.AuthUser{UserID: 8, Role: model.RoleMahasiswa}
	laboran := model.AuthUser{UserID: 2, Role: model.RoleLaboran}
	if !CanViewBooking(owner, 7) || CanViewBooking(other, 7) || !CanViewBooking(laboran, 7) {
		t.Error("aturan melihat peminjaman tidak sesuai")
	}
}
