package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
)

type RoomRepository interface {
	FindAll(ctx context.Context) ([]model.Room, error)
	FindByID(ctx context.Context, id int) (model.Room, error)
	Create(ctx context.Context, req model.RoomRequest) (model.Room, error)
	Update(ctx context.Context, id int, req model.RoomRequest) (model.Room, error)
}

type roomRepository struct{ db DBTX }

func NewRoomRepository(db DBTX) RoomRepository { return &roomRepository{db: db} }

const roomColumns = `id, code, name, building, capacity, is_active, created_at`

func scanRoom(row pgx.Row) (model.Room, error) {
	var r model.Room
	err := row.Scan(&r.ID, &r.Code, &r.Name, &r.Building, &r.Capacity, &r.IsActive, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Room{}, ErrNotFound
	}
	if pgCode(err) == "23505" {
		return model.Room{}, ErrDuplicate
	}
	return r, err
}

func (r *roomRepository) FindAll(ctx context.Context) ([]model.Room, error) {
	rows, err := r.db.Query(ctx, `SELECT `+roomColumns+` FROM rooms ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := []model.Room{}
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (r *roomRepository) FindByID(ctx context.Context, id int) (model.Room, error) {
	return scanRoom(r.db.QueryRow(ctx, `SELECT `+roomColumns+` FROM rooms WHERE id = $1`, id))
}

func (r *roomRepository) Create(ctx context.Context, req model.RoomRequest) (model.Room, error) {
	return scanRoom(r.db.QueryRow(ctx,
		`INSERT INTO rooms (code, name, building, capacity, is_active)
         VALUES ($1, $2, $3, $4, COALESCE($5, TRUE)) RETURNING `+roomColumns,
		req.Code, req.Name, req.Building, req.Capacity, req.IsActive))
}

func (r *roomRepository) Update(ctx context.Context, id int, req model.RoomRequest) (model.Room, error) {
	return scanRoom(r.db.QueryRow(ctx,
		`UPDATE rooms SET code = $2, name = $3, building = $4, capacity = $5,
		        is_active = COALESCE($6, is_active)
		 WHERE id = $1 RETURNING `+roomColumns,
		id, req.Code, req.Name, req.Building, req.Capacity, req.IsActive))
}
