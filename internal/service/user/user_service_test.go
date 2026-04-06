package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alxaxenov/ya-gophermart/internal/repo/model"
	"github.com/alxaxenov/ya-gophermart/internal/service/user/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestService_RegisterUser(t *testing.T) {
	type args struct {
		login    string
		password string
	}
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockexpectedRepo)
		args      args
		want      int
		expErr    error
		isErr     error
	}{
		{
			name: "CheckLogin error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CheckLogin(gomock.Any(), gomock.Any()).Times(1).
					Return(false, errors.New("test error"))
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			args:   args{"test", "test"},
			want:   0,
			expErr: errors.New("test error"),
			isErr:  nil,
		},
		{
			name: "Already exist",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CheckLogin(gomock.Any(), gomock.Any()).Times(1).
					Return(true, nil)
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			args:   args{"already_exist", "test"},
			want:   0,
			expErr: NewAlreadyExists("already_exist", nil),
			isErr:  nil,
		},
		{
			name: "hashPass error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CheckLogin(gomock.Any(), gomock.Any()).Times(1).
					Return(false, nil)
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			args:   args{"test", strings.Repeat("a", 73)},
			want:   0,
			expErr: nil,
			isErr:  bcrypt.ErrPasswordTooLong,
		},
		{
			name: "CreateUser error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CheckLogin(gomock.Any(), gomock.Any()).Times(1).
					Return(false, nil)
				repo.EXPECT().CreateUser(gomock.Any(), "test_login", gomock.Any()).Times(1).Return(0, errors.New("test error"))
			},
			args:   args{"test_login", "test_pass"},
			want:   0,
			expErr: errors.New("test error"),
			isErr:  nil,
		},
		{
			name: "Ok",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CheckLogin(gomock.Any(), gomock.Any()).Times(1).
					Return(false, nil)
				repo.EXPECT().CreateUser(gomock.Any(), "test_login", gomock.Any()).Times(1).Return(123, nil)
			},
			args:   args{"test_login", "test_pass"},
			want:   123,
			expErr: nil,
			isErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepoMock := mocks.NewMockexpectedRepo(ctrl)
			tt.mockSetup(userRepoMock)

			u := &Service{
				secretKey:     "",
				cookieAuthKey: "",
				expireHours:   1,
				userRepo:      userRepoMock,
			}
			got, err := u.RegisterUser(context.Background(), tt.args.login, tt.args.password)
			if tt.expErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else if tt.isErr != nil {
				assert.ErrorIs(t, err, tt.isErr, "The error should be ErrNotFound")
			} else {
				assert.NoError(t, err)
			}
			if got != tt.want {
				t.Errorf("RegisterUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_LoginUser(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockexpectedRepo)
		args      args
		want      int
		expErr    error
	}{
		{
			name: "GetUser error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetUser(gomock.Any(), "test_login").Times(1).Return(nil, errors.New("test error"))
			},
			args:   args{""},
			want:   0,
			expErr: errors.New("test error"),
		},
		{
			name: "No user",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetUser(gomock.Any(), "test_login").Times(1).Return(nil, nil)
			},
			args:   args{""},
			want:   0,
			expErr: &NotFoundError{"test_login", nil},
		},
		{
			name: "User inactive",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetUser(gomock.Any(), "test_login").Times(1).
					Return(&model.User{123, "", "", false}, nil)
			},
			args:   args{""},
			want:   0,
			expErr: &InactiveError{"test_login", nil},
		},
		{
			name: "Incorrect password",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetUser(gomock.Any(), "test_login").Times(1).
					Return(&model.User{123, "", "$2a$10$fPMLuiwmtt8VRD2E5aLOG.oZw7E/T9SJ40Yt1LRYiijrEbF.nN71G", true}, nil)
			},
			args:   args{"incorrect_password"},
			want:   0,
			expErr: IncorrectCredentialsError,
		},
		{
			name: "Ok",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetUser(gomock.Any(), "test_login").Times(1).
					Return(&model.User{123, "", "$2a$10$fPMLuiwmtt8VRD2E5aLOG.oZw7E/T9SJ40Yt1LRYiijrEbF.nN71G", true}, nil)
			},
			args:   args{"test_pass"},
			want:   123,
			expErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepoMock := mocks.NewMockexpectedRepo(ctrl)
			tt.mockSetup(userRepoMock)

			u := &Service{
				secretKey:     "",
				cookieAuthKey: "",
				expireHours:   1,
				userRepo:      userRepoMock,
			}
			got, err := u.LoginUser(context.Background(), "test_login", tt.args.password)
			if tt.expErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else {
				assert.NoError(t, err)
			}
			if got != tt.want {
				t.Errorf("LoginUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_GetCookieAuthKey(t *testing.T) {
	tests := []struct {
		name    string
		authKey string
		want    string
	}{
		{
			name:    "empty",
			authKey: "",
			want:    "",
		},
		{
			name:    "not empty",
			authKey: "qwerty",
			want:    "qwerty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Service{
				secretKey:     "",
				cookieAuthKey: tt.authKey,
				expireHours:   1,
				userRepo:      nil,
			}
			assert.Equalf(t, tt.want, u.GetCookieAuthKey(), "GetCookieAuthKey()")
		})
	}
}

func TestService_GetUserIDCtx(t *testing.T) {

	tests := []struct {
		name   string
		userID int
		want   int
		expErr error
	}{
		{
			name:   "Without ID",
			userID: 0,
			want:   0,
			expErr: IDNotFoundError,
		},
		{
			name:   "Zero",
			userID: 0,
			want:   0,
			expErr: IDNotFoundError,
		},
		{
			name:   "Ok",
			userID: 123,
			want:   123,
			expErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Service{
				secretKey:     "",
				cookieAuthKey: "",
				expireHours:   1,
				userRepo:      nil,
			}
			ctx := context.Background()
			if tt.userID != 0 {
				ctx = context.WithValue(context.Background(), UserIDKey, tt.userID)
			}
			got, err := u.GetUserIDCtx(ctx)
			if tt.expErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equalf(t, tt.want, got, "GetUserIDCtx(%v)", tt.userID)
		})
	}
}

func TestService_SetUserIDCtx(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		want   int
	}{
		{
			name:   "Ok",
			userID: 123,
			want:   123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Service{
				secretKey:     "",
				cookieAuthKey: "",
				expireHours:   1,
				userRepo:      nil,
			}
			ctx := u.SetUserIDCtx(context.Background(), tt.userID)
			resultID, ok := ctx.Value(UserIDKey).(int)
			if !ok {
				t.Errorf("SetUserIDCtx() userID not found in context")
			}
			assert.Equalf(t, tt.want, resultID, "SetUserIDCtx(%v)", tt.userID)
		})
	}
}
