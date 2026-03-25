package handler

import (
	"net/http"
	"time"

	"github.com/alxaxenov/ya-gophermart/internal/logger"
	"github.com/alxaxenov/ya-gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	GetUserBalance(w http.ResponseWriter, r *http.Request)
	AddWithdraw(w http.ResponseWriter, r *http.Request)
	GetWithdrawals(w http.ResponseWriter, r *http.Request)
	AddOrder(w http.ResponseWriter, r *http.Request)
	GetOrders(w http.ResponseWriter, r *http.Request)
}

const (
	timeoutDefault = 3 * time.Second
	timeoutBatch   = 5 * time.Second
)

type ComplexMiddleware interface {
	Use(http.Handler) http.Handler
}

func Serve(addr string, h Handler, userMiddleware ComplexMiddleware) error {
	r := chi.NewRouter()

	r.Use(middleware.WithLogging)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(userMiddleware.Use)

			r.Get("/balance", h.GetUserBalance)
			r.Post("/balance/withdraw", h.AddWithdraw)
			r.Get("/withdrawals", h.GetWithdrawals)
			r.Post("/orders", h.AddOrder)
			r.Get("/orders", h.GetOrders)
		})

	})
	//TODO timeout ручек

	logger.Logger.Info("Running server on", addr)
	return http.ListenAndServe(addr, r)
}
