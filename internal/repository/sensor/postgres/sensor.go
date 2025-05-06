package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/usecase"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorRepository struct {
	pool *pgxpool.Pool
}

func NewSensorRepository(pool *pgxpool.Pool) *SensorRepository {
	return &SensorRepository{
		pool: pool,
	}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	//goland:noinspection SqlInsertValues
	query := `INSERT INTO sensors (%s) VALUES (%s) %s RETURNING id`

	var columns []string
	var placeholders []string
	var values []interface{}
	var conflictClause string

	i := 1

	if sensor.ID != 0 {
		columns = append(columns, "id")
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, sensor.ID)
		i++
	}

	fields := []struct {
		name  string
		value any
	}{
		{"serial_number", sensor.SerialNumber},
		{"type", sensor.Type},
		{"current_state", sensor.CurrentState},
		{"description", sensor.Description},
		{"is_active", sensor.IsActive},
		{"registered_at", time.Now()},
		{"last_activity", sensor.LastActivity},
	}

	for _, field := range fields {
		columns = append(columns, field.name)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, field.value)
		i++
	}

	if sensor.ID != 0 {
		conflictClause = `ON CONFLICT (id) DO UPDATE SET
			serial_number = EXCLUDED.serial_number,
			type = EXCLUDED.type,
			current_state = EXCLUDED.current_state,
			description = EXCLUDED.description,
			is_active = EXCLUDED.is_active,
			last_activity = EXCLUDED.last_activity`
	} else {
		conflictClause = ""
	}

	finalQuery := fmt.Sprintf(query, strings.Join(columns, ", "), strings.Join(placeholders, ", "), conflictClause)

	row := r.pool.QueryRow(ctx, finalQuery, values...)
	return row.Scan(&sensor.ID)
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,
       									serial_number, 
       									type, current_state, 
       									description, is_active, 
       									registered_at, 
       									last_activity FROM sensors`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sensors []domain.Sensor

	for rows.Next() {
		sensor := &domain.Sensor{}
		if err := rows.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description, &sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity); err != nil {
			return nil, err
		}
		sensors = append(sensors, *sensor)
	}
	return sensors, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, 
       								serial_number, 
       								type, 
       								current_state, 
       								description, 
       								is_active, 
       								registered_at, 
       								last_activity FROM sensors WHERE id = $1`, id)
	sensor := &domain.Sensor{}
	if err := row.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description, &sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrSensorNotFound
		}
		return nil, err
	}
	return sensor, nil
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, 
       								serial_number, 
       								type, 
       								current_state, 
       								description, 
       								is_active, 
       								registered_at, 
       								last_activity FROM sensors WHERE serial_number = $1`, sn)
	sensor := &domain.Sensor{}
	if err := row.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description, &sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrSensorNotFound
		}
		return nil, err
	}
	return sensor, nil
}
