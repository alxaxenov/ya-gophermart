package server

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

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks
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

// Register хендлер регистрации нового пользователя
// Возможные ответы:
// 200 - пользователь успешно зарегистрирован и аутентифицирован
// 400 - неверный формат запроса (неверный content-type запроса, некорректные данные запроса)
// 409 - логин уже занят
// 500 - внутренняя ошибка сервера (необработанная ошибка бизнес логики, ошибка установки заголовка с куки пользователя)
func (h *GophermartHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != "application/json" {
		http.Error(w, "unexpected content-type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	req := model.Credentials{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("Register json decoder error", "error", err)
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

// Login аутентификация пользователя
// Возможные ответы:
// 200 - пользователь успешно аутентифицирован
// 400 - неверный формат запроса (неверный content-type запроса, некорректные данные запроса)
// 401 - неверная пара логин/пароль
// 403 - пользователь не активен
// 404 - пользователь не зарегистрирован
// 500 - внутренняя ошибка сервера (необработанная ошибка бизнес логики, ошибка установки заголовка с куки пользователя)
func (h *GophermartHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != "application/json" {
		http.Error(w, "unexpected content-type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	req := model.Credentials{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("Login json decoder error", "error", err)
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

// GetUserBalance получение текущего баланса пользователя
// Возможные ответы:
// 200 - успешная обработка запроса (возврат баланса)
// 401 - пользователь не авторизован
// 404 - сущность баланса не найдена
// 500 - внутренняя ошибка сервера (необработанная ошибка бизнес логики, ошибка конвертации в json тела ответа)
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

// AddWithdraw запрос на списание средств
// Возможные ответы:
// 200 - успешная обработка запроса
// 400 - неверный формат запроса (неверный content-type запроса, некорректные данные запроса)
// 401 - пользователь не авторизован
// 402 - на счету недостаточно средств
// 422 - неверный номер заказа
// 404 - сущность баланса не найдена
// 500 - внутренняя ошибка сервера (необработанная ошибка бизнес логики)
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
		logger.Logger.Error("AddWithdraw json decoder error", "error", err)
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

// GetWithdrawals получение информации о списаниях средств
// Возможные ответы:
// 200 - успешная обработка запроса (возврат листинга списаний)
// 204 - у пользователя нет ни одного списания
// 401 - пользователь не авторизован
// 500 - внутренняя ошибка сервера (необработанная ошибка бизнес логики)
func (h *GophermartHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userId, err := h.UserService.GetUserIDCtx(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
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

// AddOrder загрузка нового заказа в систему
// Возможные ответы:
// 200 - номер заказа уже был загружен этим пользователем (заказ уже был создан ранее для этого пользователя)
// 202 - новый номер заказа принят в обработку (новый заказ успешно создан)
// 400 - неверный формат запроса (неверный content-type запроса, некорректные данные запроса)
// 401 - пользователь не авторизован
// 409 - номер заказа уже был загружен другим пользователем (заказ уже был создан для другого пользователя)
// 422 - неверный номер заказа
// 500 - внутренняя ошибка сервера (необработанная ошибка бизнес логики)
func (h *GophermartHandler) AddOrder(w http.ResponseWriter, r *http.Request) {
	userId, err := h.UserService.GetUserIDCtx(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
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

// GetOrders получение списка загруженных заказов
// Возможные ответы:
// 200 - успешная обработка запроса (возврат листинга заказов)
// 204 - у пользователя нет зарегистрированных заказов
// 401 - пользователь не авторизован
// 500 - внутренняя ошибка сервера (необработанная ошибка бизнес логики, ошибка конвертации в json тела ответа)
func (h *GophermartHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userId, err := h.UserService.GetUserIDCtx(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
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

// addAuthCookie установка заголовка ответа Set-Cookie с кукой пользователя
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
