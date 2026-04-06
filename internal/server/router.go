package server

import (
	"net/http"
	"time"

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
	timeoutListing = 5 * time.Second
)

type ComplexMiddleware interface {
	Use(http.Handler) http.Handler
}

// initRouter конфигурация api интерфейса сервера
func initRouter(h Handler, userMiddleware ComplexMiddleware) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.WithLogging)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", timeoutHandler(h.Register, timeoutDefault, ""))
		r.Post("/login", timeoutHandler(h.Login, timeoutDefault, ""))

		r.Group(func(r chi.Router) {
			r.Use(userMiddleware.Use)

			r.Get("/balance", timeoutHandler(h.GetUserBalance, timeoutDefault, ""))
			r.Post("/balance/withdraw", timeoutHandler(h.AddWithdraw, timeoutDefault, ""))
			r.Get("/withdrawals", timeoutHandler(h.GetWithdrawals, timeoutListing, ""))
			r.Post("/orders", timeoutHandler(h.AddOrder, timeoutDefault, ""))
			r.Get("/orders", timeoutHandler(h.GetOrders, timeoutListing, ""))
		})

	})

	return r
}
