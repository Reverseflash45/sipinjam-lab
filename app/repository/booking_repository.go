package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
)

type BookingRepository interface {
	// Create menyimpan booking beserta items. Wajib dipanggil di dalam
	// transaksi (tx adalah pgx.Tx), supaya booking + items atomik.
	Create(ctx context.Context, tx DBTX, b *model.Booking) (*model.Booking, error)

	// FindByID mengambil booking beserta items. Tidak butuh transaksi.
	FindByID(ctx context.Context, id int) (*model.Booking, error)

	// UpdateStatus mengubah status booking (dipakai approve/reject/return/cancel).
	UpdateStatus(ctx context.Context, tx DBTX, id int, status model.BookingStatus,
		reviewedBy *int, note *string, returnedAt *time.Time) error

	// ListByUser mengembalikan riwayat peminjaman milik satu user.
	ListByUser(ctx context.Context, userID int) ([]model.Booking, error)

	// ListAll mengembalikan seluruh pengajuan; status nil berarti semua.
	ListAll(ctx context.Context, status *model.BookingStatus) ([]model.Booking, error)

	// CountActiveByUser menghitung booking pending/approved milik user,
	// untuk menegakkan aturan "maksimal N peminjaman aktif".
	CountActiveByUser(ctx context.Context, userID int) (int, error)

	// LockAndCheckItems mengunci baris items FOR UPDATE lalu memastikan
	// stok mencukupi untuk rentang waktu yang diminta. Mengembalikan
	// ErrItemUnavailable bila stok kurang.
	LockAndCheckItems(ctx context.Context, tx DBTX, items []model.BookingItem,
		start, end time.Time) error

	// FindApprovedByRoom mengembalikan booking approved untuk satu ruang
	// pada rentang waktu tertentu. Untuk jadwal per hari:
	// from = awal hari, to = from.Add(24*time.Hour).
	FindApprovedByRoom(ctx context.Context, roomID int, from, to time.Time) ([]model.Booking, error)

	// WithTx menjalankan fn dalam satu transaksi database. Alias dari
	// TxManager.WithTx, disediakan untuk kompatibilitas kode service 4b.
	WithTx(ctx context.Context, fn func(ctx context.Context, tx DBTX) error) error
}

type bookingRepository struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) BookingRepository {
	return &bookingRepository{pool: pool}
}

// ---------------------------------------------------------------- Create

func (r *bookingRepository) Create(ctx context.Context, tx DBTX, b *model.Booking) (*model.Booking, error) {
	const q = `
		INSERT INTO bookings (user_id, room_id, purpose, start_at, end_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := tx.QueryRow(ctx, q,
		b.UserID, b.RoomID, b.Purpose, b.StartAt, b.EndAt, b.Status,
	).Scan(&b.ID, &b.CreatedAt)

	if err != nil {
		if IsExclusionViolation(err) {
			return nil, ErrRoomConflict
		}
		if IsForeignKeyViolation(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	for i := range b.Items {
		const qi = `INSERT INTO booking_items (booking_id, item_id, quantity) VALUES ($1, $2, $3)`
		if _, err := tx.Exec(ctx, qi, b.ID, b.Items[i].ItemID, b.Items[i].Quantity); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// ------------------------------------------------------------- FindByID

func (r *bookingRepository) FindByID(ctx context.Context, id int) (*model.Booking, error) {
	const q = `
		SELECT id, user_id, room_id, purpose, start_at, end_at, status,
		       reviewed_by, review_note, returned_at, created_at
		FROM bookings WHERE id = $1`

	var b model.Booking
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&b.ID, &b.UserID, &b.RoomID, &b.Purpose, &b.StartAt, &b.EndAt,
		&b.Status, &b.ReviewedBy, &b.ReviewNote, &b.ReturnedAt, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := r.loadItems(ctx, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// ---------------------------------------------------------- UpdateStatus

func (r *bookingRepository) UpdateStatus(ctx context.Context, tx DBTX, id int,
	status model.BookingStatus, reviewedBy *int, note *string, returnedAt *time.Time) error {

	const q = `
		UPDATE bookings
		SET status = $1, reviewed_by = $2, review_note = $3, returned_at = $4
		WHERE id = $5`

	tag, err := tx.Exec(ctx, q, status, reviewedBy, note, returnedAt, id)
	if err != nil {
		if IsExclusionViolation(err) {
			return ErrRoomConflict
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ------------------------------------------------------------ ListByUser

func (r *bookingRepository) ListByUser(ctx context.Context, userID int) ([]model.Booking, error) {
	const q = `
		SELECT id, user_id, room_id, purpose, start_at, end_at, status,
		       reviewed_by, review_note, returned_at, created_at
		FROM bookings WHERE user_id = $1
		ORDER BY created_at DESC`

	return r.list(ctx, q, userID)
}

// --------------------------------------------------------------- ListAll

func (r *bookingRepository) ListAll(ctx context.Context, status *model.BookingStatus) ([]model.Booking, error) {
	if status != nil {
		const q = `
			SELECT id, user_id, room_id, purpose, start_at, end_at, status,
			       reviewed_by, review_note, returned_at, created_at
			FROM bookings WHERE status = $1
			ORDER BY created_at DESC`
		return r.list(ctx, q, *status)
	}

	const q = `
		SELECT id, user_id, room_id, purpose, start_at, end_at, status,
		       reviewed_by, review_note, returned_at, created_at
		FROM bookings ORDER BY created_at DESC`
	return r.list(ctx, q)
}

// --------------------------------------------------- CountActiveByUser

func (r *bookingRepository) CountActiveByUser(ctx context.Context, userID int) (int, error) {
	const q = `
		SELECT COUNT(*) FROM bookings
		WHERE user_id = $1 AND status IN ('pending','approved')`

	var n int
	err := r.pool.QueryRow(ctx, q, userID).Scan(&n)
	return n, err
}

// ------------------------------------------------- LockAndCheckItems

func (r *bookingRepository) LockAndCheckItems(ctx context.Context, tx DBTX,
	items []model.BookingItem, start, end time.Time) error {

	for _, it := range items {
		const lock = `SELECT stock FROM items WHERE id = $1 FOR UPDATE`
		var stock int
		if err := tx.QueryRow(ctx, lock, it.ItemID).Scan(&stock); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}

		const used = `
			SELECT COALESCE(SUM(bi.quantity), 0)
			FROM booking_items bi
			JOIN bookings b ON b.id = bi.booking_id
			WHERE bi.item_id = $1
			  AND b.status = 'approved'
			  AND tstzrange(b.start_at, b.end_at, '[)') && tstzrange($2, $3, '[)')`

		var already int
		if err := tx.QueryRow(ctx, used, it.ItemID, start, end).Scan(&already); err != nil {
			return err
		}
		if already+it.Quantity > stock {
			return ErrItemUnavailable
		}
	}
	return nil
}

// ------------------------------------------------ FindApprovedByRoom

func (r *bookingRepository) FindApprovedByRoom(ctx context.Context, roomID int, from, to time.Time) ([]model.Booking, error) {
	const q = `
		SELECT id, user_id, room_id, purpose, start_at, end_at, status,
		       reviewed_by, review_note, returned_at, created_at
		FROM bookings
		WHERE room_id = $1
		  AND status = 'approved'
		  AND tstzrange(start_at, end_at, '[)') && tstzrange($2, $3, '[)')
		ORDER BY start_at`

	return r.list(ctx, q, roomID, from, to)
}

// ------------------------------------------------------------ WithTx

func (r *bookingRepository) WithTx(ctx context.Context, fn func(ctx context.Context, tx DBTX) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ---------------------------------------------------------------- internal

func (r *bookingRepository) list(ctx context.Context, q string, args ...any) ([]model.Booking, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.RoomID, &b.Purpose, &b.StartAt,
			&b.EndAt, &b.Status, &b.ReviewedBy, &b.ReviewNote,
			&b.ReturnedAt, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Muat items untuk setiap booking (N+1, cukup untuk MVP).
	for i := range out {
		if err := r.loadItems(ctx, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (r *bookingRepository) loadItems(ctx context.Context, b *model.Booking) error {
	const q = `SELECT item_id, quantity FROM booking_items WHERE booking_id = $1`
	rows, err := r.pool.Query(ctx, q, b.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var it model.BookingItem
		if err := rows.Scan(&it.ItemID, &it.Quantity); err != nil {
			return err
		}
		b.Items = append(b.Items, it)
	}
	return rows.Err()
}