package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
)

type Sensor struct {
	repo SensorRepository
}

func NewSensor(sr SensorRepository) *Sensor {
	return &Sensor{repo: sr}
}

func (s *Sensor) RegisterSensor(ctx context.Context, sensor *domain.Sensor) (*domain.Sensor, error) {
	if sensor == nil {
		return nil, errors.New("sensor is nil")
	}
	if sensor.Type != domain.SensorTypeADC && sensor.Type != domain.SensorTypeContactClosure {
		return nil, ErrWrongSensorType
	}
	if len(sensor.SerialNumber) != 10 {
		return nil, ErrWrongSensorSerialNumber
	}

	if sens, err := s.repo.GetSensorBySerialNumber(ctx, sensor.SerialNumber); err == nil {
		return sens, nil
	} else if !errors.Is(err, ErrSensorNotFound) {
		return nil, err
	}

	if err := s.repo.SaveSensor(ctx, sensor); err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *Sensor) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	return s.repo.GetSensors(ctx)
}

func (s *Sensor) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	return s.repo.GetSensorByID(ctx, id)
}
