package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

type SensorRepository struct {
	mu         sync.RWMutex
	serialToId map[string]int64
	sensors    map[int64]*domain.Sensor
	nextID     int64
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{
		serialToId: make(map[string]int64),
		sensors:    make(map[int64]*domain.Sensor),
		nextID:     1,
	}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if sensor == nil {
		return errors.New("sensor is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		r.mu.Lock()
		defer r.mu.Unlock()
		if _, ok := r.serialToId[sensor.SerialNumber]; ok {
			return nil
		}
		sensor.ID = r.nextID
		sensor.RegisteredAt = time.Now()
		r.serialToId[sensor.SerialNumber] = r.nextID
		r.sensors[r.nextID] = sensor
		r.nextID++
	}
	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.mu.RLock()
		defer r.mu.RUnlock()
		sensors := make([]domain.Sensor, 0, len(r.sensors))
		for _, sensor := range r.sensors {
			sensors = append(sensors, *sensor)
		}
		return sensors, nil
	}
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.mu.RLock()
		defer r.mu.RUnlock()
		sensor, exists := r.sensors[id]
		if !exists {
			return nil, usecase.ErrSensorNotFound
		}
		return sensor, nil
	}
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	return r.GetSensorByID(ctx, r.serialToId[sn])
}
