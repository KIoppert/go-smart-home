package usecase

import (
	"context"
	"homework/internal/domain"
	"time"
)

type Event struct {
	eventRepo  EventRepository
	sensorRepo SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{eventRepo: er, sensorRepo: sr}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if event.Timestamp.IsZero() {
		return ErrInvalidEventTimestamp
	}

	sensor, err := e.sensorRepo.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
	if err != nil {
		return err
	}
	if sensor == nil {
		return ErrSensorNotFound
	}
	event.SensorID = sensor.ID
	if err = e.eventRepo.SaveEvent(ctx, event); err != nil {
		return err
	}
	sensor.CurrentState = event.Payload
	sensor.LastActivity = event.Timestamp
	err = e.sensorRepo.SaveSensor(ctx, sensor)
	return err
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	return e.eventRepo.GetLastEventBySensorID(ctx, id)
}

func (e *Event) GetEventsBySensorID(ctx context.Context, id int64, start, end time.Time) ([]*domain.Event, error) {
	events, err := e.eventRepo.GetEventsBySensorID(ctx, id, start, end)
	if err != nil {
		return nil, err
	}
	return events, nil
}
