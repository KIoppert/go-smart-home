package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

type EventRepository struct {
	mu     sync.Mutex
	events map[int64]map[time.Time]*domain.Event
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events: make(map[int64]map[time.Time]*domain.Event),
	}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if event == nil {
		return errors.New("event is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		r.mu.Lock()
		defer r.mu.Unlock()

		if _, exists := r.events[event.SensorID]; !exists {
			r.events[event.SensorID] = make(map[time.Time]*domain.Event)
		}
		r.events[event.SensorID][event.Timestamp] = event
		return nil
	}
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.mu.Lock()
		defer r.mu.Unlock()

		events, exists := r.events[id]
		if !exists || len(events) == 0 {
			return nil, usecase.ErrEventNotFound
		}

		var latestEvent *domain.Event
		for _, event := range events {
			if latestEvent == nil || event.Timestamp.After(latestEvent.Timestamp) {
				latestEvent = event
			}
		}
		return latestEvent, nil
	}
}

func (r *EventRepository) GetEventsBySensorID(ctx context.Context, id int64, start, end time.Time) ([]*domain.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.mu.Lock()
		defer r.mu.Unlock()
		if start.IsZero() || end.IsZero() || start.After(end) {
			return nil, usecase.ErrInvalidEventTimestamp
		}

		events, exists := r.events[id]
		if !exists {
			return nil, usecase.ErrSensorNotFound
		}

		var result []*domain.Event
		for _, event := range events {
			if event.Timestamp.After(start) && event.Timestamp.Before(end) {
				result = append(result, event)
			}
		}
		return result, nil
	}
}
