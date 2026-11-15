package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxManager menjalankan fungsi dalam satu transaksi database. Fungsi
// menerima DBTX yang bisa dipakai repository mana pun — bisa pgx.Tx
// (di dalam transaksi) atau pool (di luar).
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, tx DBTX) error) error
}

type txManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) TxManager {
	return &txManager{pool: pool}
}

func (m *txManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context, tx DBTX) error) error {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	// Rollback aman dipanggil setelah commit: pgx mengabaikannya
	// bila transaksi sudah selesai (committed atau rolled back).
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// IsExclusionViolation mendeteksi pelanggaran EXCLUDE constraint
// (SQLSTATE 23P01). Dipakai repository untuk menerjemahkan bentrok
// jadwal ruang menjadi ErrRoomConflict → response 409 ROOM_CONFLICT.
func IsExclusionViolation(err error) bool {
	return pgCode(err) == "23P01"
}

// IsForeignKeyViolation mendeteksi pelanggaran FK (SQLSTATE 23503).
func IsForeignKeyViolation(err error) bool {
	return pgCode(err) == "23503"
}

// IsCheckViolation mendeteksi pelanggaran CHECK constraint (SQLSTATE 23514).
func IsCheckViolation(err error) bool {
	return pgCode(err) == "23514"
}