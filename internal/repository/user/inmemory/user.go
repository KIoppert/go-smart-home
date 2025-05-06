package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
)

type UserRepository struct {
	mu     sync.RWMutex
	users  map[int64]*domain.User
	nextID int64
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:  make(map[int64]*domain.User),
		nextID: 1,
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		r.mu.Lock()
		defer r.mu.Unlock()

		user.ID = r.nextID
		r.nextID++
		r.users[user.ID] = user
		return nil
	}
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.mu.RLock()
		defer r.mu.RUnlock()

		user, exists := r.users[id]
		if !exists {
			return nil, usecase.ErrUserNotFound
		}
		return user, nil
	}
}
