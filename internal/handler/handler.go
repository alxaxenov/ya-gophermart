package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/alxaxenov/ya-gophermart/internal/logger"
	"github.com/alxaxenov/ya-gophermart/internal/model"
	"github.com/alxaxenov/ya-gophermart/internal/service/gophermart"
	"github.com/alxaxenov/ya-gophermart/internal/service/user"
)

type GophermartHandler struct {
	UserService       UserService
	GophermartService GophermartService
}

func NewGophermartHandler(userService UserService, gophermartService GophermartService) *GophermartHandler {
	return &GophermartHandler{
		userService,
		gophermartService,
	}
}

type UserService interface {
	RegisterUser(context.Context, string, string) (int, error)
	LoginUser(context.Context, string, string) (int, error)
	GetCookieAuthKey() string
	BuildTokenString(int) (string, error)
	GetUserIDCtx(context.Context) (int, error)
}

type GophermartService interface {
	GetBalance(context.Context, int) (*model.UserBalanceResponse, error)
	AddWithdraw(context.Context, int, string, float64) error
	GetWithdrawals(context.Context, int) (*[]model.GetWithdrawResponse, error)
	CreateOrder(context.Context, int, string) (bool, error)
	GetOrders(context.Context, int) (*[]model.GetOrdersResponse, error)
}

func (h *GophermartHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != "application/json" {
		http.Error(w, "unexpected content-type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	req := model.Credentials{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Infof("Register json decoder error %s", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := model.Validate(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := h.UserService.RegisterUser(r.Context(), req.Login, req.Password)
	if err != nil {
		var exist *user.AlreadyExistsError
		if errors.As(err, &exist) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		logger.Logger.Error("Register unexpected error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = h.addAuthCookie(userID, &w)
	if err != nil {
		logger.Logger.Error("Cookie set error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *GophermartHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != "application/json" {
		http.Error(w, "unexpected content-type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	req := model.Credentials{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Infof("Login json decoder error %s", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := model.Validate(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := h.UserService.LoginUser(r.Context(), req.Login, req.Password)
	if err != nil {
		var inactive *user.InactiveError
		var notFound *user.NotFoundError
		switch {
		case errors.As(err, &inactive):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, user.IncorrectCredentialsError):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case errors.As(err, &notFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			logger.Logger.Error("Login unexpected error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	err = h.addAuthCookie(userID, &w)
	if err != nil {
		logger.Logger.Error("Cookie set error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *GophermartHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userId, err := h.UserService.GetUserIDCtx(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	balance, err := h.GophermartService.GetBalance(r.Context(), userId)
	if err != nil {
		if errors.Is(err, gophermart.BalanceNotFoundError) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		logger.Logger.Error("GetUserBalance unexpected error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	respData, err := json.Marshal(balance)
	if err != nil {
		logger.Logger.Error("GetUserBalance failed to marshal response", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

func (h *GophermartHandler) AddWithdraw(w http.ResponseWriter, r *http.Request) {
	userId, err := h.UserService.GetUserIDCtx(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if r.Header.Get("content-type") != "application/json" {
		http.Error(w, "unexpected content-type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	req := model.AddWithdrawRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Infof("AddWithdraw json decoder error %s", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := model.Validate(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.GophermartService.AddWithdraw(r.Context(), userId, req.Order, req.Sum)
	if err != nil {
		var status int
		switch {
		case errors.Is(err, gophermart.InsufficientFundsError):
			status = http.StatusPaymentRequired
		case errors.Is(err, gophermart.OrderNumberIncorrectError):
			status = http.StatusUnprocessableEntity
		case errors.Is(err, gophermart.BalanceNotFoundError):
			status = http.StatusNotFound
		default:
			logger.Logger.Error("AddWithdraw unexpected error", "error", err)
			status = http.StatusInternalServerError
		}
		var text string
		if status != http.StatusInternalServerError {
			text = err.Error()
		} else {
			text = http.StatusText(status)
		}
		http.Error(w, text, status)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *GophermartHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userId, err := h.UserService.GetUserIDCtx(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
	withdrawals, err := h.GophermartService.GetWithdrawals(r.Context(), userId)
	if err != nil {
		logger.Logger.Error("GetWithdrawals unexpected error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	status := http.StatusOK
	var respData []byte
	if len(*withdrawals) == 0 {
		status = http.StatusNoContent
		respData = []byte(http.StatusText(http.StatusNoContent))
	} else {
		respData, err = json.Marshal(model.WithdrawalsResponse(*withdrawals))
		if err != nil {
			logger.Logger.Error("GetWithdrawals response marshal error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(respData)

}

func (h *GophermartHandler) AddOrder(w http.ResponseWriter, r *http.Request) {
	userId, err := h.UserService.GetUserIDCtx(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

	if r.Header.Get("content-type") != "text/plain" {
		http.Error(w, "unexpected content-type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	orderNum, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Logger.Error("AddOrder read body error", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	created, err := h.GophermartService.CreateOrder(r.Context(), userId, string(orderNum))
	if err != nil {
		var status int
		switch {
		case errors.Is(err, gophermart.OrderNumberIncorrectError):
			status = http.StatusUnprocessableEntity
		case errors.Is(err, gophermart.AnotherUserError):
			status = http.StatusConflict
		default:
			logger.Logger.Error("AddOrder unexpected error", "error", err)
			status = http.StatusInternalServerError
		}
		var text string
		if status != http.StatusInternalServerError {
			text = err.Error()
		} else {
			text = http.StatusText(status)
		}
		http.Error(w, text, status)
		return
	}

	var status int
	if created {
		status = http.StatusAccepted
	} else {
		status = http.StatusOK
	}
	w.WriteHeader(status)
}

func (h *GophermartHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userId, err := h.UserService.GetUserIDCtx(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

	orders, err := h.GophermartService.GetOrders(r.Context(), userId)
	if err != nil {
		logger.Logger.Error("GetOrders unexpected error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	status := http.StatusOK
	var respData []byte
	if len(*orders) == 0 {
		status = http.StatusNoContent
		respData = []byte(http.StatusText(http.StatusNoContent))
	} else {
		respData, err = json.Marshal(model.OrdersResponse(*orders))
		if err != nil {
			logger.Logger.Error("GetOrders response marshal error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(respData)
}

func (h *GophermartHandler) addAuthCookie(ID int, w *http.ResponseWriter) error {
	token, err := h.UserService.BuildTokenString(ID)
	if err != nil {
		return err
	}
	http.SetCookie(*w, &http.Cookie{
		Name:     h.UserService.GetCookieAuthKey(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	return nil
}
