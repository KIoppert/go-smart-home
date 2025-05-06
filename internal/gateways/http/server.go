package http

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	host           string
	port           uint16
	router         *gin.Engine
	shutdownRouter func() error
}

type UseCases struct {
	Event  *usecase.Event
	Sensor *usecase.Sensor
	User   *usecase.User
}

func NewServer(useCases UseCases, options ...func(*Server)) *Server {
	r := gin.Default()

	ws := NewWebSocketHandler(useCases)
	setupRouter(r, useCases, ws)

	s := &Server{router: r, host: "localhost", port: 8080, shutdownRouter: ws.Shutdown}
	for _, o := range options {
		o(s)
	}

	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(ctx context.Context, cancel context.CancelFunc) error {
	serv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s.router,
	}
	defer cancel()

	go func() {
		_ = serv.ListenAndServe()
	}()

	<-ctx.Done()
	errs := make([]error, 1)
	if err := serv.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	err := s.shutdownRouter()
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
