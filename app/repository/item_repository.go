package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
)

type ItemRepository interface {
	WithTx(tx DBTX) ItemRepository

	FindAll(ctx context.Context) ([]model.Item, error)
	FindByID(ctx context.Context, id int) (model.Item, error)
	Create(ctx context.Context, req model.ItemRequest) (model.Item, error)
	Update(ctx context.Context, id int, req model.ItemRequest) (model.Item, error)

	// LockForUpdate mengunci baris alat supaya dua persetujuan yang
	// bersamaan tidak sama-sama mengira stoknya masih cukup.
	LockForUpdate(ctx context.Context, ids []int) error
	// BorrowedQuantity menjumlahkan alat yang sedang dipinjam (disetujui,
	// belum dikembalikan) pada rentang waktu yang beririsan.
	BorrowedQuantity(ctx context.Context, itemID int, start, end time.Time) (int, error)
}

type itemRepository struct{ db DBTX }

func NewItemRepository(db DBTX) ItemRepository { return &itemRepository{db: db} }

func (r *itemRepository) WithTx(tx DBTX) ItemRepository { return &itemRepository{db: tx} }

const itemColumns = `id, code, name, category, stock, is_active, created_at`

func scanItem(row pgx.Row) (model.Item, error) {
	var it model.Item
	err := row.Scan(&it.ID, &it.Code, &it.Name, &it.Category, &it.Stock, &it.IsActive, &it.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Item{}, ErrNotFound
	}
	if pgCode(err) == "23505" {
		return model.Item{}, ErrDuplicate
	}
	return it, err
}

func (r *itemRepository) FindAll(ctx context.Context) ([]model.Item, error) {
	rows, err := r.db.Query(ctx, `SELECT `+itemColumns+` FROM items ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *itemRepository) FindByID(ctx context.Context, id int) (model.Item, error) {
	return scanItem(r.db.QueryRow(ctx, `SELECT `+itemColumns+` FROM items WHERE id = $1`, id))
}

func (r *itemRepository) Create(ctx context.Context, req model.ItemRequest) (model.Item, error) {
	return scanItem(r.db.QueryRow(ctx,
		`INSERT INTO items (code, name, category, stock) VALUES ($1, $2, $3, $4) RETURNING `+itemColumns,
		req.Code, req.Name, req.Category, req.Stock))
}

func (r *itemRepository) Update(ctx context.Context, id int, req model.ItemRequest) (model.Item, error) {
	return scanItem(r.db.QueryRow(ctx,
		`UPDATE items SET code = $2, name = $3, category = $4, stock = $5,
		        is_active = COALESCE($6, is_active)
		 WHERE id = $1 RETURNING `+itemColumns,
		id, req.Code, req.Name, req.Category, req.Stock, req.IsActive))
}

func (r *itemRepository) LockForUpdate(ctx context.Context, ids []int) error {
	_, err := r.db.Exec(ctx, `SELECT id FROM items WHERE id = ANY($1) ORDER BY id FOR UPDATE`, ids)
	return err
}

func (r *itemRepository) BorrowedQuantity(ctx context.Context, itemID int, start, end time.Time) (int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(bi.quantity), 0)
		 FROM booking_items bi JOIN bookings b ON b.id = bi.booking_id
		 WHERE bi.item_id = $1 AND b.status = 'approved'
		   AND tstzrange(b.start_at, b.end_at) && tstzrange($2, $3)`,
		itemID, start, end).Scan(&total)
	return total, err
}
