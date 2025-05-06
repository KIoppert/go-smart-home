package http

import (
	"errors"
	"homework/internal/domain"
	"homework/internal/models"
	"homework/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

func makeSens(sens *domain.Sensor) models.Sensor {
	sensorType := string(sens.Type)
	lastActivity := strfmt.DateTime(sens.LastActivity)
	registeredAt := strfmt.DateTime(sens.RegisteredAt)
	sensor := models.Sensor{
		ID:           &sens.ID,
		Description:  &sens.Description,
		SerialNumber: &sens.SerialNumber,
		Type:         &sensorType,
		CurrentState: &sens.CurrentState,
		IsActive:     &sens.IsActive,
		LastActivity: &lastActivity,
		RegisteredAt: &registeredAt,
	}
	return sensor
}

func checkHeader(ctx *gin.Context, err error) {
	if ctx.GetHeader("Accept") != "application/json" {
		str := err.Error()
		ctx.AbortWithStatusJSON(http.StatusNotAcceptable, models.Error{Reason: &str})
	}
}

func getSensor(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		checkHeader(ctx, errors.New("accept header must be application/json"))

		if ctx.IsAborted() {
			return
		}

		sens, err := us.Sensor.GetSensors(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		sensors := make([]models.Sensor, 0, len(sens))
		for i, sensor := range sens {
			sensors[i] = makeSens(&sensor)
		}
		ctx.JSON(http.StatusOK, sensors)
	}
}

func headSensor(_ UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		checkHeader(ctx, errors.New("accept header must be application/json"))
		if ctx.IsAborted() {
			return
		}
		ctx.Header("Content-Length", "1")
		ctx.Status(http.StatusOK)
	}
}

func postSensor(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		toCreate := &models.SensorToCreate{}

		validateSensor(ctx, toCreate)

		if ctx.IsAborted() {
			return
		}

		sensor, err := us.Sensor.RegisterSensor(ctx, &domain.Sensor{
			Description:  *toCreate.Description,
			SerialNumber: *toCreate.SerialNumber,
			Type:         domain.SensorType(*toCreate.Type),
			IsActive:     *toCreate.IsActive,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, makeSens(sensor))
	}
}

func checkContentType(ctx *gin.Context) {
	if ctx.GetHeader("content-type") != "application/json" {
		ctx.AbortWithStatus(http.StatusUnsupportedMediaType)
	}
}

func validateSensor(ctx *gin.Context, toCreate *models.SensorToCreate) {
	checkContentType(ctx)
	if err := ctx.ShouldBindJSON(toCreate); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	if err := toCreate.Validate(nil); err != nil {
		ctx.AbortWithStatus(http.StatusUnprocessableEntity)
	}
}

func commonGet(ctx *gin.Context, us UseCases) *domain.Sensor {
	checkHeader(ctx, errors.New("accept header must be application/json"))

	sensorID, err := strconv.Atoi(ctx.Param("sensor_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, models.Error{Reason: swag.String("sensor_id must be a number")})
	}

	if ctx.IsAborted() {
		return nil
	}
	var sensor *domain.Sensor
	if sensor, err = us.Sensor.GetSensorByID(ctx, int64(sensorID)); errors.Is(err, usecase.ErrSensorNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, models.Error{Reason: swag.String("sensor not found")})
		return nil
	} else if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, models.Error{Reason: swag.String("internal server error")})
		return nil
	}
	return sensor
}

func getSensorByID(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensor := commonGet(ctx, us)
		if !ctx.IsAborted() {
			ctx.JSON(http.StatusOK, makeSens(sensor))
		}
	}
}

func headSensorByID(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_ = commonGet(ctx, us)
		if !ctx.IsAborted() {
			ctx.Header("Content-Length", "1")
			ctx.Status(http.StatusOK)
		}
	}
}

func subscribe(us UseCases, wsh *WebSocketHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensorID, err := strconv.Atoi(ctx.Param("sensor_id"))
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, models.Error{Reason: swag.String("sensor_id is required")})
			return
		}
		sensor, err := us.Sensor.GetSensorByID(ctx, int64(sensorID))
		if err != nil {
			if errors.Is(err, usecase.ErrSensorNotFound) {
				ctx.JSON(http.StatusNotFound, models.Error{Reason: swag.String("sensor not found")})
				return
			}
			ctx.JSON(http.StatusInternalServerError, models.Error{Reason: swag.String(err.Error())})
			return
		}
		err = wsh.Handle(ctx, sensor.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.Error{Reason: swag.String(err.Error())})
		}
	}
}

func getHistory(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensor := commonGet(ctx, us)
		if ctx.IsAborted() {
			return
		}

		start := ctx.Query("start_date")
		end := ctx.Query("end_date")

		if start == "" || end == "" {
			ctx.JSON(http.StatusBadRequest, models.Error{Reason: swag.String("start and end query parameters are required")})
			return
		}

		startTime, err := time.Parse(time.RFC1123, start)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.Error{Reason: swag.String("invalid start_date format")})
			return
		}
		endTime, err := time.Parse(time.RFC1123, end)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.Error{Reason: swag.String("invalid end_date format")})
			return
		}
		history, err := us.Event.GetEventsBySensorID(ctx, sensor.ID, startTime, endTime)
		if err != nil {
			if errors.Is(err, usecase.ErrEventNotFound) {
				ctx.JSON(http.StatusNotFound, models.Error{Reason: swag.String("event not found")})
				return
			}
			ctx.JSON(http.StatusInternalServerError, models.Error{Reason: swag.String(err.Error())})
			return
		}
		answer := make([]models.HistoryOfEvents, len(history))
		for i, event := range history {
			Timestamp := strfmt.DateTime(event.Timestamp)
			answer[i] = models.HistoryOfEvents{
				Payload:   &event.Payload,
				Timestamp: &Timestamp,
			}
		}
		ctx.JSON(http.StatusOK, answer)
	}
}
