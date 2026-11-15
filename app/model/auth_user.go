package model

// Role adalah tipe kuat untuk role pengguna. Nilainya harus sama dengan
// CHECK constraint di tabel users: admin, laboran, dosen, mahasiswa.
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleLaboran   Role = "laboran"
	RoleDosen     Role = "dosen"
	RoleMahasiswa Role = "mahasiswa"
)

// AuthUser adalah identitas minimal yang disimpan middleware RequireAuth
// ke Fiber Locals: user_id dan role.
type AuthUser struct {
	ID   int  `json:"id"`
	Role Role `json:"role"`
}

// IsStaff mengembalikan true untuk admin dan laboran — dua role yang
// boleh menyetujui, menolak, dan mencatat pengembalian peminjaman.
func (u AuthUser) IsStaff() bool {
	return u.Role == RoleAdmin || u.Role == RoleLaboran
}

// IsPeminjam mengembalikan true untuk dosen dan mahasiswa — dua role
// yang boleh mengajukan peminjaman.
func (u AuthUser) IsPeminjam() bool {
	return u.Role == RoleDosen || u.Role == RoleMahasiswa
}

// CanManageBookings mengembalikan true bila user boleh approve/reject/return.
func (u AuthUser) CanManageBookings() bool {
	return u.IsStaff()
}