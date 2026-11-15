# SiPinjam Lab

Sistem peminjaman ruang dan alat laboratorium Fakultas Vokasi Universitas Airlangga. Mahasiswa dan dosen
mengajukan peminjaman lewat API, laboran menyetujui atau menolak, dan sistem menjaga agar ruang tidak
dipinjam dua kali pada jam yang sama dan alat tidak dipinjam melebihi stok.

Tugas 13 Praktikum Pemrograman Backend Lanjut, D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

## Stack

Go, Fiber v2, PostgreSQL (pgx v5), JWT HS256 (gofiber/contrib/jwt), bcrypt, go-playground/validator.

## Struktur (Clean Architecture sesuai modul perkuliahan)

```
main.go               perakitan dependency: repository -> service -> route
config/               env dan NewApp (Fiber + satu error handler terpusat)
database/             koneksi PostgreSQL
route/                peta URL dan hak akses
middleware/           middleware global, RequireAuth, RequireRole
app/model/            struct data, request, dan response
app/repository/       akses database (interface + implementasi PostgreSQL)
app/service/          logika bisnis; method menerima *fiber.Ctx langsung
helper/               error, response, validator, identitas user login
migrations/           skema dan data awal
```

Tidak ada layer controller maupun handler. Route memanggil method service secara langsung, service memanggil
repository lewat interface, dan repository adalah satu-satunya bagian yang menulis SQL.

## Menjalankan

```powershell
psql -U postgres -c "CREATE DATABASE sipinjam_lab;"
psql -U postgres -d sipinjam_lab -f migrations/000002_create_users.up.sql
psql -U postgres -d sipinjam_lab -f migrations/002_rooms_items.sql
psql -U postgres -d sipinjam_lab -f migrations/003_bookings.sql
psql -U postgres -d sipinjam_lab -f migrations/004_seed.sql
Copy-Item .env.example .env      # isi DB_PASSWORD dan JWT_SECRET
go mod tidy
go test ./...
go run .
```

Login memakai NIM atau NIP (`identity_no`) dan password. Akun awal (password `Rahasia123`): NIP
`198001012005011001` (admin) dan `198502022010012002` (laboran). Mahasiswa dan dosen mendaftar sendiri lewat
`POST /api/v1/auth/register`, dan response register maupun login langsung berisi token JWT.

## Endpoint

| Method | Path | Akses |
|---|---|---|
| GET | /api/v1/health | publik |
| POST | /api/v1/auth/register | publik (role mahasiswa atau dosen) |
| POST | /api/v1/auth/login | publik |
| GET | /api/v1/auth/me | login |
| GET | /api/v1/rooms, /rooms/:id, /rooms/:id/schedule?date= | login |
| POST, PUT | /api/v1/rooms, /rooms/:id | admin |
| GET | /api/v1/items, /items/:id | login |
| POST, PUT | /api/v1/items, /items/:id | admin, laboran |
| POST | /api/v1/bookings | mahasiswa, dosen |
| GET | /api/v1/bookings/mine | login |
| GET | /api/v1/bookings?status= | admin, laboran |
| GET | /api/v1/bookings/:id | pemilik, admin, laboran |
| PATCH | /api/v1/bookings/:id/approve, /reject, /return | admin, laboran |
| PATCH | /api/v1/bookings/:id/cancel | pemilik (mahasiswa, dosen) |

## Aturan bisnis utama

- Peminjaman maksimal 8 jam, dimulai di masa depan, paling jauh 30 hari ke depan.
- Peminjaman aktif (pending dan approved) maksimal 3 untuk mahasiswa dan 5 untuk dosen.
- Status: `pending` ke `approved`, `rejected`, atau `cancelled`; `approved` ke `returned`.
- Ruang yang sama tidak boleh punya dua peminjaman disetujui yang waktunya beririsan. Dijaga oleh
  `EXCLUDE USING gist` di PostgreSQL, sehingga tetap aman walaupun dua laboran menyetujui bersamaan.
- Saat disetujui, jumlah alat yang sedang dipinjam pada rentang waktu yang sama ditambah permintaan baru tidak
  boleh melebihi stok. Baris alat dikunci `FOR UPDATE` di dalam transaction.
