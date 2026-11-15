package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX dipenuhi oleh *pgxpool.Pool maupun pgx.Tx, sehingga repository yang
// sama bisa dipakai di luar maupun di dalam transaction.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var (
	ErrNotFound     = errors.New("data tidak ditemukan")
	ErrDuplicate    = errors.New("data sudah ada")
	ErrRoomConflict = errors.New("ruang sudah dipinjam pada waktu tersebut")
	ErrItemUnavailable = errors.New("stok alat tidak mencukupi") 
)

// pgCode mengembalikan kode error PostgreSQL, misalnya 23505 untuk UNIQUE.
func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
