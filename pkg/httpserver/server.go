package httpserver

import (
	"context"
	"net/http"
	"shortener/pkg/handler"
	"time"
)

type Server struct {
	server *http.Server
}

func New(handler *handler.Handler, address string) *Server {
	return &Server{server: &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
