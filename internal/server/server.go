package server

import (
	"context"
	"net/http"

	"github.com/alxaxenov/ya-gophermart/internal/logger"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	addr        string
	innerServer *http.Server
	router      *chi.Mux
}

func NewServer(addr string, h Handler, userMiddleware ComplexMiddleware) *Server {
	r := initRouter(h, userMiddleware)
	srv := http.Server{
		Addr:    addr,
		Handler: r,
	}
	return &Server{addr: addr, innerServer: &srv, router: r}
}

func (s *Server) Start() error {
	logger.Logger.Info("Running server", "address", s.addr)
	return s.innerServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	logger.Logger.Info("Stopping server")
	return s.innerServer.Shutdown(ctx)
}
