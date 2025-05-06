package inmemory

import (
	"context"
	"homework/internal/domain"
	"sync"
)

type SensorOwnerRepository struct {
	mu      sync.RWMutex
	sensors map[int64][]domain.SensorOwner
}

func NewSensorOwnerRepository() *SensorOwnerRepository {
	return &SensorOwnerRepository{
		sensors: make(map[int64][]domain.SensorOwner),
	}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		r.mu.Lock()
		defer r.mu.Unlock()
		r.sensors[sensorOwner.UserID] = append(r.sensors[sensorOwner.UserID], sensorOwner)
	}
	return nil
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.mu.RLock()
		defer r.mu.RUnlock()
		if sensors, exists := r.sensors[userID]; exists {
			return sensors, nil
		}
	}
	return make([]domain.SensorOwner, 0), nil
}
