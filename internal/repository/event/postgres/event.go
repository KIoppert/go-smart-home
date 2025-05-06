package postgres

import (
	"context"
	"errors"
	"homework/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEventNotFound = errors.New("event not found")

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		pool,
	}
}

func (r *EventRepository) GetEventsBySensorID(ctx context.Context, id int64, start, end time.Time) ([]*domain.Event, error) {
	rows, err := r.pool.Query(ctx, `SELECT timestamp, sensor_serial_number, sensor_id, payload FROM events WHERE sensor_id = $1 AND timestamp BETWEEN $2 AND $3`, id, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []*domain.Event
	for rows.Next() {
		event := &domain.Event{}
		if err := rows.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO events (timestamp, sensor_serial_number, sensor_id, payload) VALUES ($1, $2, $3, $4)`, event.Timestamp, event.SensorSerialNumber, event.SensorID, event.Payload)
	if err != nil {
		return err
	}
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	row := r.pool.QueryRow(ctx, `SELECT timestamp, sensor_serial_number, sensor_id, payload FROM events WHERE sensor_id = $1 ORDER BY timestamp DESC LIMIT 1`, id)
	event := &domain.Event{}
	if err := row.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	return event, nil
}
