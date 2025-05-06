package http

import (
	"errors"
	"homework/internal/models"
	"homework/internal/usecase"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/swag"
)

type WebSocketHandler struct {
	useCases   UseCases
	websockets mapset.Set[*websocket.Conn]
	mutex      sync.Mutex
}

func NewWebSocketHandler(useCases UseCases) *WebSocketHandler {
	return &WebSocketHandler{
		useCases:   useCases,
		websockets: mapset.NewSet[*websocket.Conn](),
		mutex:      sync.Mutex{},
	}
}

func (h *WebSocketHandler) Handle(c *gin.Context, id int64) error {
	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.Error{Reason: swag.String("websocket accept error")})
	}
	h.mutex.Lock()
	h.websockets.Add(conn)
	h.mutex.Unlock()

	ticker := time.NewTicker(time.Millisecond * 500)
	ctx := conn.CloseRead(c)

	select {
	case <-ctx.Done():
		break
	case <-ticker.C:
		event, err := h.useCases.Event.GetLastEventBySensorID(c, id)
		if errors.Is(err, usecase.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, models.Error{Reason: swag.String("not enough events")})
			break
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, models.Error{Reason: swag.String("websocket accept error")})
			break
		}
		err = wsjson.Write(c, conn, event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Error{Reason: swag.String("failed to write message")})
			break
		}
	}
	h.mutex.Lock()
	h.websockets.Remove(conn)
	h.mutex.Unlock()
	_ = conn.Close(websocket.StatusNormalClosure, "connection closed")
	ticker.Stop()
	return nil
}

func (h *WebSocketHandler) Shutdown() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	for conn := range h.websockets.Iter() {
		err := conn.Close(websocket.StatusNormalClosure, "server shutting down")
		if err != nil {
			log.Printf("failed to close websocket: %v", err)
		}
	}
	return nil
}
