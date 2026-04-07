package server

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alxaxenov/ya-gophermart/internal/model"
	"github.com/alxaxenov/ya-gophermart/internal/server/mocks"
	"github.com/alxaxenov/ya-gophermart/internal/service/gophermart"
	"github.com/alxaxenov/ya-gophermart/internal/service/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGophermartHandler_Register(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		contentType string
		want        want
		body        io.Reader
		mockSetup   func(sr *mocks.MockUserService)
	}{
		{
			name:        "Wrong content-type",
			contentType: "text/plain",
			want:        want{http.StatusBadRequest},
			body:        nil,
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().RegisterUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "Decode error",
			contentType: "application/json",
			want:        want{http.StatusBadRequest},
			body:        strings.NewReader(``),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().RegisterUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "Validate error zero values",
			contentType: "application/json",
			want:        want{http.StatusBadRequest},
			body:        strings.NewReader(`{"login": "", "password": ""}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().RegisterUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "AlreadyExistsError",
			contentType: "application/json",
			want:        want{http.StatusConflict},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().RegisterUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(0, &user.AlreadyExistsError{})
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "another error",
			contentType: "application/json",
			want:        want{http.StatusInternalServerError},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().RegisterUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(0, errors.New("test_error"))
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "addAuthCookie error",
			contentType: "application/json",
			want:        want{http.StatusInternalServerError},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().RegisterUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(123, nil)
				sr.EXPECT().BuildTokenString(123).Times(1).Return("", errors.New("test_error"))
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "Ok",
			contentType: "application/json",
			want:        want{http.StatusOK},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().RegisterUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(123, nil)
				sr.EXPECT().BuildTokenString(123).Times(1).Return("test_token", nil)
				sr.EXPECT().GetCookieAuthKey().Times(1).Return("test_auth_key")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userServiceMock := mocks.NewMockUserService(ctrl)
			tt.mockSetup(userServiceMock)
			h := &GophermartHandler{
				UserService:       userServiceMock,
				GophermartService: nil,
			}
			r := httptest.NewRequest(http.MethodPost, "/api/user/register", tt.body)
			r.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			h.Register(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
		})
	}
}

func TestGophermartHandler_Login(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		contentType string
		want        want
		body        io.Reader
		mockSetup   func(sr *mocks.MockUserService)
	}{
		{
			name:        "Wrong content-type",
			contentType: "text/plain",
			want:        want{http.StatusBadRequest},
			body:        nil,
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "Decode error",
			contentType: "application/json",
			want:        want{http.StatusBadRequest},
			body:        strings.NewReader(``),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "Validate error zero values",
			contentType: "application/json",
			want:        want{http.StatusBadRequest},
			body:        strings.NewReader(`{"login": "", "password": ""}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "InactiveError",
			contentType: "application/json",
			want:        want{http.StatusForbidden},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(0, &user.InactiveError{})
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "IncorrectCredentialsError",
			contentType: "application/json",
			want:        want{http.StatusUnauthorized},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(0, user.IncorrectCredentialsError)
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "NotFoundError",
			contentType: "application/json",
			want:        want{http.StatusNotFound},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(0, &user.NotFoundError{})
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "Another error",
			contentType: "application/json",
			want:        want{http.StatusInternalServerError},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(0, errors.New("test_error"))
				sr.EXPECT().BuildTokenString(gomock.Any()).Times(0)
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},

		{
			name:        "addAuthCookie error",
			contentType: "application/json",
			want:        want{http.StatusInternalServerError},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(123, nil)
				sr.EXPECT().BuildTokenString(123).Times(1).Return("", errors.New("test_error"))
				sr.EXPECT().GetCookieAuthKey().Times(0)
			},
		},
		{
			name:        "Ok",
			contentType: "application/json",
			want:        want{http.StatusOK},
			body:        strings.NewReader(`{"login": "test_login", "password": "test_password"}`),
			mockSetup: func(sr *mocks.MockUserService) {
				sr.EXPECT().LoginUser(gomock.Any(), "test_login", "test_password").Times(1).
					Return(123, nil)
				sr.EXPECT().BuildTokenString(123).Times(1).Return("test_token", nil)
				sr.EXPECT().GetCookieAuthKey().Times(1).Return("test_auth_key")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userServiceMock := mocks.NewMockUserService(ctrl)
			tt.mockSetup(userServiceMock)
			h := &GophermartHandler{
				UserService:       userServiceMock,
				GophermartService: nil,
			}
			r := httptest.NewRequest(http.MethodPost, "/api/user/login", tt.body)
			r.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			h.Login(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
		})
	}
}

func TestGophermartHandler_GetUserBalance(t *testing.T) {
	type want struct {
		statusCode int
		body       *model.UserBalanceResponse
	}
	tests := []struct {
		name      string
		mockSetup func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService)
		want      want
	}{
		{
			name: "GetUserIDCtx error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(0, errors.New("test_error")).Times(1)
				gsr.EXPECT().GetBalance(gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{http.StatusUnauthorized, nil},
		},
		{
			name: "BalanceNotFoundError",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetBalance(gomock.Any(), 123).Times(1).
					Return(nil, gophermart.BalanceNotFoundError)
			},
			want: want{http.StatusNotFound, nil},
		},
		{
			name: "Another error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetBalance(gomock.Any(), 123).Times(1).
					Return(nil, errors.New("test_error"))
			},
			want: want{http.StatusInternalServerError, nil},
		},
		{
			name: "Marshal error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetBalance(gomock.Any(), 123).Times(1).
					Return(&model.UserBalanceResponse{math.NaN(), math.NaN()}, nil)
			},
			want: want{http.StatusInternalServerError, nil},
		},
		{
			name: "Ok",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetBalance(gomock.Any(), 123).Times(1).
					Return(&model.UserBalanceResponse{25.55, 42.24}, nil)
			},
			want: want{http.StatusOK, &model.UserBalanceResponse{25.55, 42.24}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userServiceMock := mocks.NewMockUserService(ctrl)
			gophermartServiceMock := mocks.NewMockGophermartService(ctrl)
			tt.mockSetup(userServiceMock, gophermartServiceMock)
			h := &GophermartHandler{
				UserService:       userServiceMock,
				GophermartService: gophermartServiceMock,
			}

			r := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			w := httptest.NewRecorder()
			h.GetUserBalance(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var respBody model.UserBalanceResponse
				err := json.Unmarshal(w.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, *tt.want.body, respBody)
			}
		})
	}
}

func TestGophermartHandler_AddWithdraw(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		mockSetup   func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService)
		contentType string
		body        io.Reader
		want        want
	}{
		{
			name: "GetUserIDCtx error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(0, errors.New("test_error")).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			contentType: "application/json",
			body:        nil,
			want:        want{http.StatusUnauthorized},
		},
		{
			name: "Wrong content-type",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			contentType: "text/plain",
			body:        nil,
			want:        want{http.StatusBadRequest},
		},
		{
			name: "Decode error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			contentType: "application/json",
			body:        strings.NewReader(``),
			want:        want{http.StatusBadRequest},
		},
		{
			name: "Validate error zero values",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			contentType: "application/json",
			body:        strings.NewReader(`{"order": "", "sum": 0}`),
			want:        want{http.StatusBadRequest},
		},
		{
			name: "InsufficientFundsError",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), 123, "1234", 25.55).Times(1).
					Return(gophermart.InsufficientFundsError)
			},
			contentType: "application/json",
			body:        strings.NewReader(`{"order": "1234", "sum": 25.55}`),
			want:        want{http.StatusPaymentRequired},
		},
		{
			name: "OrderNumberIncorrectError",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), 123, "1234", 25.55).Times(1).
					Return(gophermart.OrderNumberIncorrectError)
			},
			contentType: "application/json",
			body:        strings.NewReader(`{"order": "1234", "sum": 25.55}`),
			want:        want{http.StatusUnprocessableEntity},
		},
		{
			name: "BalanceNotFoundError",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), 123, "1234", 25.55).Times(1).
					Return(gophermart.BalanceNotFoundError)
			},
			contentType: "application/json",
			body:        strings.NewReader(`{"order": "1234", "sum": 25.55}`),
			want:        want{http.StatusNotFound},
		},
		{
			name: "Another error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), 123, "1234", 25.55).Times(1).
					Return(errors.New("test_error"))
			},
			contentType: "application/json",
			body:        strings.NewReader(`{"order": "1234", "sum": 25.55}`),
			want:        want{http.StatusInternalServerError},
		},
		{
			name: "Ok",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().AddWithdraw(gomock.Any(), 123, "1234", 25.55).Times(1).
					Return(nil)
			},
			contentType: "application/json",
			body:        strings.NewReader(`{"order": "1234", "sum": 25.55}`),
			want:        want{http.StatusOK},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userServiceMock := mocks.NewMockUserService(ctrl)
			gophermartServiceMock := mocks.NewMockGophermartService(ctrl)
			tt.mockSetup(userServiceMock, gophermartServiceMock)
			h := &GophermartHandler{
				UserService:       userServiceMock,
				GophermartService: gophermartServiceMock,
			}

			r := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", tt.body)
			r.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			h.AddWithdraw(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
		})
	}
}

func TestGophermartHandler_GetWithdrawals(t *testing.T) {
	type want struct {
		statusCode int
		body       *model.WithdrawalsResponse
	}
	tests := []struct {
		name      string
		mockSetup func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService)
		want      want
	}{
		{
			name: "UserId not found",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(0, errors.New("test error")).Times(1)
				gsr.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{http.StatusUnauthorized, nil},
		},
		{
			name: "GetWithdrawals error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetWithdrawals(gomock.Any(), 123).Times(1).
					Return(nil, errors.New("test error"))
			},
			want: want{http.StatusInternalServerError, nil},
		},
		{
			name: "GetWithdrawals empty",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetWithdrawals(gomock.Any(), 123).Times(1).
					Return(&[]model.GetWithdrawResponse{}, nil)
			},
			want: want{http.StatusNoContent, nil},
		},
		{
			name: "GetWithdrawals marshal error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetWithdrawals(gomock.Any(), 123).Times(1).
					Return(
						&[]model.GetWithdrawResponse{
							{"12345", math.NaN(), "2024-06-25T21:07:45-04:00"},
						},
						nil,
					)
			},
			want: want{http.StatusInternalServerError, nil},
		},
		{
			name: "Ok",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetWithdrawals(gomock.Any(), 123).Times(1).
					Return(
						&[]model.GetWithdrawResponse{
							{"12345", 25.55, "2024-06-25T21:07:45-04:00"},
							{"54321", 42.24, "2009-08-13T18:14:30-04:00"},
						},
						nil,
					)
			},
			want: want{
				http.StatusOK,
				&model.WithdrawalsResponse{
					{"12345", 25.55, "2024-06-25T21:07:45-04:00"},
					{"54321", 42.24, "2009-08-13T18:14:30-04:00"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userServiceMock := mocks.NewMockUserService(ctrl)
			gophermartServiceMock := mocks.NewMockGophermartService(ctrl)
			tt.mockSetup(userServiceMock, gophermartServiceMock)
			h := &GophermartHandler{
				UserService:       userServiceMock,
				GophermartService: gophermartServiceMock,
			}

			r := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			w := httptest.NewRecorder()
			h.GetWithdrawals(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var respBody model.WithdrawalsResponse
				err := json.Unmarshal(w.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, *tt.want.body, respBody)
			}
		})
	}
}

type errorReader struct{}

func (errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("forced read error")
}

func (errorReader) Close() error {
	return nil
}

func TestGophermartHandler_AddOrder(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		mockSetup   func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService)
		contentType string
		body        io.Reader
		want        want
	}{
		{
			name: "GetUserIDCtx error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(0, errors.New("test_error")).Times(1)
				gsr.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			contentType: "text/plain",
			body:        nil,
			want:        want{http.StatusUnauthorized},
		},
		{
			name: "Wrong content-type",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			contentType: "application/json",
			body:        nil,
			want:        want{http.StatusBadRequest},
		},
		{
			name: "Read error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			contentType: "text/plain",
			body:        errorReader{},
			want:        want{http.StatusBadRequest},
		},
		{
			name: "OrderNumberIncorrectError",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().CreateOrder(gomock.Any(), 123, "12345").Times(1).
					Return(false, gophermart.OrderNumberIncorrectError)
			},
			contentType: "text/plain",
			body:        strings.NewReader(`12345`),
			want:        want{http.StatusUnprocessableEntity},
		},
		{
			name: "AnotherUserError",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().CreateOrder(gomock.Any(), 123, "12345").Times(1).
					Return(false, gophermart.AnotherUserError)
			},
			contentType: "text/plain",
			body:        strings.NewReader(`12345`),
			want:        want{http.StatusConflict},
		},
		{
			name: "Another error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().CreateOrder(gomock.Any(), 123, "12345").Times(1).
					Return(false, errors.New("test_error"))
			},
			contentType: "text/plain",
			body:        strings.NewReader(`12345`),
			want:        want{http.StatusInternalServerError},
		},
		{
			name: "not created",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().CreateOrder(gomock.Any(), 123, "12345").Times(1).
					Return(false, nil)
			},
			contentType: "text/plain",
			body:        strings.NewReader(`12345`),
			want:        want{http.StatusOK},
		},
		{
			name: "created",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().CreateOrder(gomock.Any(), 123, "12345").Times(1).
					Return(true, nil)
			},
			contentType: "text/plain",
			body:        strings.NewReader(`12345`),
			want:        want{http.StatusAccepted},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userServiceMock := mocks.NewMockUserService(ctrl)
			gophermartServiceMock := mocks.NewMockGophermartService(ctrl)
			tt.mockSetup(userServiceMock, gophermartServiceMock)
			h := &GophermartHandler{
				UserService:       userServiceMock,
				GophermartService: gophermartServiceMock,
			}

			r := httptest.NewRequest(http.MethodPost, "/api/user/orders", tt.body)
			r.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			h.AddOrder(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
		})
	}
}

func TestGophermartHandler_GetOrders(t *testing.T) {
	type want struct {
		statusCode int
		body       *model.OrdersResponse
	}
	tests := []struct {
		name      string
		mockSetup func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService)
		want      want
	}{
		{
			name: "UserId not found",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(0, errors.New("test error")).Times(1)
				gsr.EXPECT().GetOrders(gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{http.StatusUnauthorized, nil},
		},
		{
			name: "GetOrders error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetOrders(gomock.Any(), 123).Times(1).
					Return(nil, errors.New("test error"))
			},
			want: want{http.StatusInternalServerError, nil},
		},
		{
			name: "no content",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetOrders(gomock.Any(), 123).Times(1).
					Return(&[]model.GetOrdersResponse{}, nil)
			},
			want: want{http.StatusNoContent, nil},
		},
		{
			name: "Marshal error",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetOrders(gomock.Any(), 123).Times(1).
					Return(
						&[]model.GetOrdersResponse{
							{"12345", "PROCESSED", math.NaN(), "2024-06-25T21:07:45-04:00"},
						},
						nil,
					)
			},
			want: want{http.StatusInternalServerError, nil},
		},
		{
			name: "Ok",
			mockSetup: func(usr *mocks.MockUserService, gsr *mocks.MockGophermartService) {
				usr.EXPECT().GetUserIDCtx(gomock.Any()).Return(123, nil).Times(1)
				gsr.EXPECT().GetOrders(gomock.Any(), 123).Times(1).
					Return(
						&[]model.GetOrdersResponse{
							{"12345", "PROCESSED", 25.55, "2024-06-25T21:07:45-04:00"},
							{"54321", "PROCESSING", 42.24, "2009-08-13T18:14:30-04:00"},
						},
						nil,
					)
			},
			want: want{
				http.StatusOK,
				&model.OrdersResponse{
					{"12345", "PROCESSED", 25.55, "2024-06-25T21:07:45-04:00"},
					{"54321", "PROCESSING", 42.24, "2009-08-13T18:14:30-04:00"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userServiceMock := mocks.NewMockUserService(ctrl)
			gophermartServiceMock := mocks.NewMockGophermartService(ctrl)
			tt.mockSetup(userServiceMock, gophermartServiceMock)
			h := &GophermartHandler{
				UserService:       userServiceMock,
				GophermartService: gophermartServiceMock,
			}
			r := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			w := httptest.NewRecorder()
			h.GetOrders(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var respBody model.OrdersResponse
				err := json.Unmarshal(w.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, *tt.want.body, respBody)
			}
		})
	}
}
