package postgres

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	row := r.pool.QueryRow(ctx, `INSERT INTO users (name) VALUES ($1) RETURNING id`, user.Name)
	return row.Scan(&user.ID)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, name FROM users WHERE id = $1`, id)
	user := &domain.User{}
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
