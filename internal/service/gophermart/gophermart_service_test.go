package gophermart

import (
	"context"
	"reflect"
	"testing"
	"time"

	"errors"

	"database/sql"

	dbMocks "github.com/alxaxenov/ya-gophermart/internal/config/db/mocks"
	"github.com/alxaxenov/ya-gophermart/internal/model"
	intModel "github.com/alxaxenov/ya-gophermart/internal/model"
	repoModel "github.com/alxaxenov/ya-gophermart/internal/repo/model"
	"github.com/alxaxenov/ya-gophermart/internal/service/gophermart/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestService_GetBalance(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockexpectedRepo)
		want      *model.UserBalanceResponse
		expErr    error
	}{
		{
			name: "Repo error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetBalance(gomock.Any(), 123).
					Times(1).Return(nil, errors.New("repo error"))
			},
			want:   nil,
			expErr: errors.New("repo error"),
		},
		{
			name: "Balance not found",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetBalance(gomock.Any(), 123).
					Times(1).Return(nil, nil)
			},
			want:   nil,
			expErr: BalanceNotFoundError,
		},
		{
			name: "Ok",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetBalance(gomock.Any(), 123).
					Times(1).Return(&repoModel.Balance{12, 13, 55.06, 64.78}, nil)
			},
			want:   &intModel.UserBalanceResponse{55.06, 64.78},
			expErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			gophRepoMock := mocks.NewMockexpectedRepo(ctrl)
			tt.mockSetup(gophRepoMock)

			g := &Service{GophermartRepo: gophRepoMock}
			got, err := g.GetBalance(context.Background(), 123)
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else {
				assert.NoError(t, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_CreateOrder(t *testing.T) {
	type args struct {
		userID   int
		orderNum string
	}
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockexpectedRepo)
		args      args
		want      bool
		expErr    error
	}{
		{
			name: "Incorrect order number",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				repo.EXPECT().GetOrderByNumber(gomock.Any(), gomock.Any()).Times(0)
			},
			args:   args{123, ""},
			want:   false,
			expErr: OrderNumberIncorrectError,
		},
		{
			name: "Create order error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CreateOrder(gomock.Any(), 123, "18").Times(1).
					Return(false, errors.New("proxy error"))
				repo.EXPECT().GetOrderByNumber(gomock.Any(), gomock.Any()).Times(0)
			},
			args:   args{123, "18"},
			want:   false,
			expErr: errors.New("proxy error"),
		},
		{
			name: "Order created",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CreateOrder(gomock.Any(), 123, "18").Times(1).
					Return(true, nil)
				repo.EXPECT().GetOrderByNumber(gomock.Any(), gomock.Any()).Times(0)
			},
			args:   args{123, "18"},
			want:   true,
			expErr: nil,
		},
		{
			name: "Get order error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CreateOrder(gomock.Any(), 123, "26").Times(1).
					Return(false, nil)
				repo.EXPECT().GetOrderByNumber(gomock.Any(), "26").Times(1).
					Return(nil, errors.New("proxy error"))
			},
			args:   args{123, "26"},
			want:   false,
			expErr: errors.New("proxy error"),
		},
		{
			name: "Exist order another user",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CreateOrder(gomock.Any(), 123, "26").Times(1).
					Return(false, nil)
				repo.EXPECT().GetOrderByNumber(gomock.Any(), "26").Times(1).
					Return(&repoModel.Order{1, 321}, nil)
			},
			args:   args{123, "26"},
			want:   false,
			expErr: AnotherUserError,
		},
		{
			name: "Exist order current user",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().CreateOrder(gomock.Any(), 123, "26").Times(1).
					Return(false, nil)
				repo.EXPECT().GetOrderByNumber(gomock.Any(), "26").Times(1).
					Return(&repoModel.Order{1, 123}, nil)
			},
			args:   args{123, "26"},
			want:   false,
			expErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			gophRepoMock := mocks.NewMockexpectedRepo(ctrl)
			tt.mockSetup(gophRepoMock)

			g := &Service{GophermartRepo: gophRepoMock}
			got, err := g.CreateOrder(context.Background(), tt.args.userID, tt.args.orderNum)
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetOrders(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockexpectedRepo)
		want      *[]model.GetOrdersResponse
		expErr    error
	}{
		{
			name: "GetOrders error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetOrders(gomock.Any(), 123).Times(1).Return(nil, errors.New("proxy error"))
			},
			want:   nil,
			expErr: errors.New("proxy error"),
		},
		{
			name: "No orders",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetOrders(gomock.Any(), 123).Times(1).Return(&[]repoModel.OrderForList{}, nil)
			},
			want:   &[]model.GetOrdersResponse{},
			expErr: nil,
		},
		{
			name: "Ok",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetOrders(gomock.Any(), 123).Times(1).
					Return(&[]repoModel.OrderForList{
						{"18", "PROCESSING", sql.NullFloat64{0, false}, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
						{"26", "PROCESSED", sql.NullFloat64{25.55, true}, time.Date(2024, time.September, 27, 16, 35, 20, 0, time.UTC)},
					},
						nil,
					)
			},
			want: &[]model.GetOrdersResponse{
				{"18", "PROCESSING", 0, "2009-11-10T23:00:00Z"},
				{"26", "PROCESSED", 25.55, "2024-09-27T16:35:20Z"},
			},
			expErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			gophRepoMock := mocks.NewMockexpectedRepo(ctrl)
			tt.mockSetup(gophRepoMock)

			g := &Service{GophermartRepo: gophRepoMock}
			got, err := g.GetOrders(context.Background(), 123)
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else {
				assert.NoError(t, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_GetWithdrawals(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockexpectedRepo)
		want      *[]model.GetWithdrawResponse
		expErr    error
	}{
		{
			name: "GetUserWithdrawals error",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetUserWithdrawals(gomock.Any(), 123).Times(1).Return(nil, errors.New("proxy error"))
			},
			want:   nil,
			expErr: errors.New("proxy error"),
		},
		{
			name: "No withdrawals",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetUserWithdrawals(gomock.Any(), 123).Times(1).Return(&[]repoModel.Withdraw{}, nil)
			},
			want:   &[]model.GetWithdrawResponse{},
			expErr: nil,
		},
		{
			name: "Ok",
			mockSetup: func(repo *mocks.MockexpectedRepo) {
				repo.EXPECT().GetUserWithdrawals(gomock.Any(), 123).Times(1).Return(
					&[]repoModel.Withdraw{
						{"18", 42.24, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
						{"26", 5.05, time.Date(2024, time.September, 27, 16, 35, 20, 0, time.UTC)},
					},
					nil,
				)
			},
			want: &[]model.GetWithdrawResponse{
				{"18", 42.24, "2009-11-10T23:00:00Z"},
				{"26", 5.05, "2024-09-27T16:35:20Z"},
			},
			expErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			gophRepoMock := mocks.NewMockexpectedRepo(ctrl)
			tt.mockSetup(gophRepoMock)

			g := &Service{GophermartRepo: gophRepoMock}
			got, err := g.GetWithdrawals(context.Background(), 123)
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else {
				assert.NoError(t, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_AddWithdraw(t *testing.T) {
	type args struct {
		order string
		sum   float64
	}
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockexpectedRepo, tx *dbMocks.MockTx)
		args      args
		expErr    error
	}{
		{
			name: "Incorrect order number",
			mockSetup: func(repo *mocks.MockexpectedRepo, tx *dbMocks.MockTx) {
				repo.EXPECT().GetBalanceTx(gomock.Any(), gomock.Any()).Times(0)
				repo.EXPECT().AddWithdrawTx(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				tx.EXPECT().Rollback().Times(0)
			},
			args:   args{"", 10},
			expErr: OrderNumberIncorrectError,
		},
		{
			name: "GetBalanceTx error",
			mockSetup: func(repo *mocks.MockexpectedRepo, tx *dbMocks.MockTx) {
				repo.EXPECT().GetBalanceTx(gomock.Any(), 123).Times(1).Return(nil, nil, errors.New("proxy error"))
				repo.EXPECT().AddWithdrawTx(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				tx.EXPECT().Rollback().Times(0)
			},
			args:   args{"18", 10},
			expErr: errors.New("proxy error"),
		},
		{
			name: "Balance not found",
			mockSetup: func(repo *mocks.MockexpectedRepo, tx *dbMocks.MockTx) {
				repo.EXPECT().GetBalanceTx(gomock.Any(), 123).Times(1).Return(nil, tx, nil)
				repo.EXPECT().AddWithdrawTx(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				tx.EXPECT().Rollback().Times(1)
			},
			args:   args{"18", 10},
			expErr: BalanceNotFoundError,
		},
		{
			name: "Insufficient funds",
			mockSetup: func(repo *mocks.MockexpectedRepo, tx *dbMocks.MockTx) {
				repo.EXPECT().GetBalanceTx(gomock.Any(), 123).Times(1).Return(
					&repoModel.Balance{1, 123, 5, 5},
					tx,
					nil,
				)
				repo.EXPECT().AddWithdrawTx(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				tx.EXPECT().Rollback().Times(1)
			},
			args:   args{"18", 10},
			expErr: InsufficientFundsError,
		},
		{
			name: "AddWithdrawTx error",
			mockSetup: func(repo *mocks.MockexpectedRepo, tx *dbMocks.MockTx) {
				repo.EXPECT().GetBalanceTx(gomock.Any(), 123).Times(1).Return(
					&repoModel.Balance{1, 123, 10, 5},
					tx,
					nil,
				)
				repo.EXPECT().AddWithdrawTx(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("proxy error"))
				tx.EXPECT().Rollback().Times(1)
			},
			args:   args{"18", 10},
			expErr: errors.New("proxy error"),
		},
		{
			name: "Ok",
			mockSetup: func(repo *mocks.MockexpectedRepo, tx *dbMocks.MockTx) {
				repo.EXPECT().GetBalanceTx(gomock.Any(), 123).Times(1).Return(
					&repoModel.Balance{52, 123, 10, 5},
					tx,
					nil,
				)
				repo.EXPECT().AddWithdrawTx(
					gomock.Any(),
					gomock.Any(),
					&repoModel.AddWithdraw{
						BalanceID:    52,
						UserID:       123,
						Order:        "18",
						Sum:          10,
						NewCurrent:   0,
						NewWithdrawn: 15,
					},
				).Times(1).Return(errors.New("proxy error"))
				tx.EXPECT().Rollback().Times(1)
			},
			args:   args{"18", 10},
			expErr: errors.New("proxy error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			gophRepoMock := mocks.NewMockexpectedRepo(ctrl)
			txMock := dbMocks.NewMockTx(ctrl)
			tt.mockSetup(gophRepoMock, txMock)

			g := &Service{GophermartRepo: gophRepoMock}
			err := g.AddWithdraw(context.Background(), 123, tt.args.order, tt.args.sum)
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
