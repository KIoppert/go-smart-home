package inmemory

import (
	"context"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventRepository_SaveEvent(t *testing.T) {
	t.Run("err, event is nil", func(t *testing.T) {
		er := NewEventRepository()
		err := er.SaveEvent(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := er.SaveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		err := er.SaveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		event := &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: "0123456789",
			Payload:            0,
		}

		err := er.SaveEvent(ctx, event)
		assert.NoError(t, err)

		actualEvent, err := er.GetLastEventBySensorID(ctx, event.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, event.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, event.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, event.Payload, actualEvent.Payload)
	})

	t.Run("ok, collision test", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := sync.WaitGroup{}
		var lastEvent domain.Event
		for i := 0; i < 1000; i++ {
			event := &domain.Event{
				Timestamp:          time.Now(),
				SensorSerialNumber: "0123456789",
				Payload:            0,
			}
			lastEvent = *event
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NoError(t, er.SaveEvent(ctx, event))
			}()
		}

		wg.Wait()

		actualEvent, err := er.GetLastEventBySensorID(ctx, lastEvent.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, lastEvent.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, lastEvent.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, lastEvent.Payload, actualEvent.Payload)
	})
}

func TestEventRepository_GetLastEventBySensorID(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := er.GetLastEventBySensorID(ctx, 0)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := er.GetLastEventBySensorID(ctx, 0)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("fail, event not found", func(t *testing.T) {
		er := NewEventRepository()
		_, err := er.GetLastEventBySensorID(context.Background(), 234)
		assert.ErrorIs(t, err, usecase.ErrEventNotFound)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorID := int64(12345)
		var lastEvent *domain.Event
		for i := 0; i < 10; i++ {
			lastEvent = &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   0,
			}
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, lastEvent))
		}

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  54321,
				Payload:   0,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		actualEvent, err := er.GetLastEventBySensorID(ctx, lastEvent.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, lastEvent.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, lastEvent.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, lastEvent.Payload, actualEvent.Payload)
	})
}

func TestEventRepository_GetEventsBySensorID(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := er.GetEventsBySensorID(ctx, 0, time.Time{}, time.Time{})
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := er.GetEventsBySensorID(ctx, 0, time.Time{}, time.Time{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sensorID := int64(1234567890)
		var lastEvent *domain.Event
		for i := 0; i < 10; i++ {
			lastEvent = &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   int64(100 + i),
			}
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, lastEvent))
		}

		actualEvents, err := er.GetEventsBySensorID(ctx, sensorID, time.Now().Add(-time.Second), time.Now())
		assert.NoError(t, err)
		assert.Equal(t, len(actualEvents), 10)
	})
}

func FuzzGetEventsBySensorID(f *testing.F) {
	start := time.Now().Add(-10 * time.Hour).Format(time.RFC3339)
	end := time.Now().Add(10 * time.Hour).Format(time.RFC3339)
	f.Add(int64(1), start, end)
	f.Add(int64(1), "invalid-date", "2023-01-02T0033:00:00Z")
	f.Add(int64(0), "invalid-date", "2023-01-02T00:00:00Z")
	f.Add(int64(-1), "2023-01-01T00:00:00Z", "invalid-date")

	er := NewEventRepository()

	ctx := context.Background()
	sensorID := int64(1)
	for i := 0; i < 10; i++ {
		event := &domain.Event{
			Timestamp: time.Now(),
			SensorID:  sensorID,
			Payload:   int64(i),
		}
		_ = er.SaveEvent(ctx, event)
	}

	f.Fuzz(func(t *testing.T, id int64, startDate, endDate string) {
		ctx := context.Background()

		start, err1 := time.Parse(time.RFC3339, startDate)
		end, err2 := time.Parse(time.RFC3339, endDate)

		if err1 != nil || err2 != nil || start.IsZero() || end.IsZero() || start.After(end) {
			_, err := er.GetEventsBySensorID(ctx, id, start, end)
			assert.Error(t, err)
			return
		}

		_, err := er.GetEventsBySensorID(ctx, id, start, end)
		if id == sensorID {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	})
}

func TestEventRepository_GetEventsBySensorID_TableDriven(t *testing.T) {
	er := NewEventRepository()
	ctx := context.Background()

	sensorID := int64(1234567890)
	for i := 0; i < 5; i++ {
		event := &domain.Event{
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			SensorID:  sensorID,
			Payload:   int64(i),
		}
		_ = er.SaveEvent(ctx, event)
	}

	tests := []struct {
		name      string
		sensorID  int64
		startTime time.Time
		endTime   time.Time
		wantCount int
		wantErr   error
	}{
		{
			name:      "valid range, events exist",
			sensorID:  sensorID,
			startTime: time.Now().Add(-time.Minute),
			endTime:   time.Now().Add(10 * time.Minute),
			wantCount: 5,
			wantErr:   nil,
		},
		{
			name:      "no events in range",
			sensorID:  sensorID,
			startTime: time.Now().Add(-10 * time.Minute),
			endTime:   time.Now().Add(-5 * time.Minute),
			wantCount: 0,
			wantErr:   nil,
		},
		{
			name:      "invalid sensor ID",
			sensorID:  -1,
			startTime: time.Now().Add(-time.Minute),
			endTime:   time.Now().Add(10 * time.Minute),
			wantCount: 0,
			wantErr:   usecase.ErrSensorNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := er.GetEventsBySensorID(ctx, tt.sensorID, tt.startTime, tt.endTime)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, events, tt.wantCount)
		})
	}
}
