package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
)

type User struct {
	userRepo   UserRepository
	sorRepo    SensorOwnerRepository
	sensorRepo SensorRepository
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{
		userRepo:   ur,
		sorRepo:    sor,
		sensorRepo: sr,
	}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}
	if user.Name == "" {
		return nil, ErrInvalidUserName
	}
	if err := u.userRepo.SaveUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return err
	}
	if _, err := u.sensorRepo.GetSensorByID(ctx, sensorID); err != nil {
		return err
	}
	err := u.sorRepo.SaveSensorOwner(ctx, domain.SensorOwner{UserID: userID, SensorID: sensorID})
	if err != nil {
		return err
	}
	return nil
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	sensorOwners, err := u.sorRepo.GetSensorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sensors := make([]domain.Sensor, 0, len(sensorOwners))
	for _, so := range sensorOwners {
		sensor, err := u.sensorRepo.GetSensorByID(ctx, so.SensorID)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, *sensor)
	}
	return sensors, nil
}
