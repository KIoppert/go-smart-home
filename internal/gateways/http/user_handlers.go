package http

import (
	"errors"
	"homework/internal/domain"
	"homework/internal/models"
	"homework/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/swag"
)

func postUser(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		toCreate := &models.UserToCreate{}

		validate(ctx, toCreate)

		if ctx.IsAborted() {
			return
		}

		user, err := us.User.RegisterUser(ctx, &domain.User{Name: *toCreate.Name})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, models.User{
			ID:   &user.ID,
			Name: &user.Name,
		})
	}
}

func commonGetUserSensors(ctx *gin.Context, us UseCases) []domain.Sensor {
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, models.Error{Reason: swag.String("user_id is required")})
		return nil
	}

	if ctx.GetHeader("Accept") != "application/json" {
		ctx.AbortWithStatus(http.StatusNotAcceptable)
		return nil
	}

	sensors, err := us.User.GetUserSensors(ctx, int64(userID))
	if errors.Is(err, usecase.ErrUserNotFound) {
		ctx.JSON(http.StatusNotFound, models.Error{Reason: swag.String("user not found")})
		return nil
	} else if errors.Is(err, usecase.ErrSensorNotFound) {
		ctx.JSON(http.StatusNotFound, models.Error{Reason: swag.String("sensor not found")})
		return nil
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Error{Reason: swag.String(err.Error())})
		return nil
	}

	return sensors
}

func getUserSensors(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensors := commonGetUserSensors(ctx, us)
		if !ctx.IsAborted() {
			ctx.JSON(http.StatusOK, sensors)
		}
	}
}

func headUserSensors(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sens := commonGetUserSensors(ctx, us)
		if !ctx.IsAborted() {
			ctx.Header("Content-Length", strconv.Itoa(len(sens)))
			ctx.Status(http.StatusOK)
		}
	}
}

func postUserSensors(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, err := strconv.Atoi(ctx.Param("user_id"))
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, models.Error{Reason: swag.String("user_id is required")})
			return
		}

		var sensor models.SensorToUserBinding
		validate(ctx, &sensor)

		if ctx.IsAborted() {
			return
		}

		err = us.User.AttachSensorToUser(ctx, int64(userID), *sensor.SensorID)
		if errors.Is(err, usecase.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, models.Error{Reason: swag.String("user not found")})
			return
		} else if errors.Is(err, usecase.ErrSensorNotFound) {
			ctx.JSON(http.StatusNotFound, models.Error{Reason: swag.String("sensor not found")})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.Error{Reason: swag.String(err.Error())})
			return
		}

		ctx.Status(http.StatusCreated)
	}
}
