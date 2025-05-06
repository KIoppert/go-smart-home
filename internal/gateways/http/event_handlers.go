package http

import (
	"homework/internal/domain"
	"homework/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func postEvent(us UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		toCreate := &models.SensorEvent{}

		validate(ctx, toCreate)

		if ctx.IsAborted() {
			return
		}

		err := us.Event.ReceiveEvent(ctx, &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: *toCreate.SensorSerialNumber,
			Payload:            *toCreate.Payload,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.Status(http.StatusCreated)
	}
}
