package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxManager menjalankan fungsi dalam satu transaksi database.
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, tx DBTX) error) error
	WithTx(ctx context.Context, fn func(ctx context.Context, tx DBTX) error) error
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
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// WithTx alias WithinTransaction.
func (m *txManager) WithTx(ctx context.Context, fn func(ctx context.Context, tx DBTX) error) error {
	return m.WithinTransaction(ctx, fn)
}

// IsExclusionViolation mendeteksi pelanggaran EXCLUDE constraint (SQLSTATE 23P01).
func IsExclusionViolation(err error) bool { return pgCode(err) == "23P01" }

// IsForeignKeyViolation mendeteksi pelanggaran FK (SQLSTATE 23503).
func IsForeignKeyViolation(err error) bool { return pgCode(err) == "23503" }

// IsCheckViolation mendeteksi pelanggaran CHECK constraint (SQLSTATE 23514).
func IsCheckViolation(err error) bool { return pgCode(err) == "23514" }