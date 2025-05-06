package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/strfmt"
)

type Validatable interface {
	Validate(formats strfmt.Registry) error
}

func validate(ctx *gin.Context, toCreate Validatable) {
	if ctx.GetHeader("content-type") != "application/json" {
		ctx.AbortWithStatus(http.StatusUnsupportedMediaType)
	}
	if err := ctx.ShouldBindJSON(toCreate); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	if err := toCreate.Validate(nil); err != nil {
		ctx.AbortWithStatus(http.StatusUnprocessableEntity)
	}
}

func optionsHandler(methods ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Allow", strings.Join(methods, ","))
		c.AbortWithStatus(http.StatusNoContent)
	}
}

func setupRouter(r *gin.Engine, us UseCases, wsh *WebSocketHandler) {
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.POST("/users", postUser(us))
	r.OPTIONS("/users", optionsHandler(http.MethodPost, http.MethodOptions))

	r.GET("/sensors", getSensor(us))
	r.HEAD("/sensors", headSensor(us))
	r.POST("/sensors", postSensor(us))
	r.OPTIONS("/sensors", optionsHandler(http.MethodHead, http.MethodGet, http.MethodPost, http.MethodOptions))

	r.GET("/sensors/:sensor_id/events", subscribe(us, wsh))
	r.GET("/sensors/:sensor_id", getSensorByID(us))
	r.HEAD("/sensors/:sensor_id", headSensorByID(us))
	r.OPTIONS("/sensors/:sensor_id", optionsHandler(http.MethodHead, http.MethodGet, http.MethodPost, http.MethodOptions))
	r.GET("/sensors/:sensor_id/history", getHistory(us))

	r.GET("/users/:user_id/sensors", getUserSensors(us))
	r.HEAD("/users/:user_id/sensors", headUserSensors(us))
	r.POST("/users/:user_id/sensors", postUserSensors(us))
	r.OPTIONS("/users/:user_id/sensors", optionsHandler(http.MethodHead, http.MethodGet, http.MethodPost, http.MethodOptions))

	r.POST("/events", postEvent(us))
	r.OPTIONS("/events", optionsHandler(http.MethodPost, http.MethodOptions))

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/users") ||
			strings.HasPrefix(c.Request.URL.Path, "/sensors") ||
			strings.HasPrefix(c.Request.URL.Path, "/events") {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
		}
	})
}
