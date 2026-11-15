package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/Reverseflash45/sipinjam-lab/app/model"
)

type UserRepository interface {
	FindByIdentityNo(ctx context.Context, identityNo string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	Create(ctx context.Context, u *model.User) (*model.User, error)
}

type userRepository struct {
	db DBTX
}

func NewUserRepository(db DBTX) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByIdentityNo(ctx context.Context, identityNo string) (*model.User, error) {
	query := `SELECT id, name, email, identity_no, password_hash, role, is_active, created_at FROM users WHERE identity_no = $1`
	return r.scanOne(ctx, query, identityNo)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, name, email, identity_no, password_hash, role, is_active, created_at FROM users WHERE email = $1`
	return r.scanOne(ctx, query, email)
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	query := `SELECT id, name, email, identity_no, password_hash, role, is_active, created_at FROM users WHERE id = $1`
	return r.scanOne(ctx, query, id)
}

func (r *userRepository) scanOne(ctx context.Context, query string, arg any) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, query, arg).Scan(
		&u.ID, &u.Name, &u.Email, &u.IdentityNo,
		&u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	query := `
		INSERT INTO users (name, email, identity_no, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, email, identity_no, password_hash, role, is_active, created_at`

	err := r.db.QueryRow(ctx, query,
		u.Name, u.Email, u.IdentityNo, u.PasswordHash, u.Role,
	).Scan(
		&u.ID, &u.Name, &u.Email, &u.IdentityNo,
		&u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if pgCode(err) == "23505" {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return u, nil
}