package gophermart

import (
	"context"
	"time"

	"github.com/alxaxenov/ya-gophermart/internal/config/db"
	"github.com/alxaxenov/ya-gophermart/internal/model"
	repo_model "github.com/alxaxenov/ya-gophermart/internal/repo/model"
	"github.com/alxaxenov/ya-gophermart/internal/utils"
)

type Service struct {
	GophermartRepo expectedRepo
}

type expectedRepo interface {
	GetBalance(context.Context, int) (*repo_model.Balance, error)
	GetBalanceTx(context.Context, int) (*repo_model.Balance, db.Tx, error)
	AddWithdrawTx(context.Context, db.Tx, *repo_model.AddWithdraw) error
	GetUserWithdrawals(context.Context, int) (*[]repo_model.Withdraw, error)
	CreateOrder(context.Context, int, string) (bool, error)
	GetOrderByNumber(context.Context, string) (*repo_model.Order, error)
	GetOrders(ctx context.Context, userID int) (*[]repo_model.OrderForList, error)
}

func (g *Service) GetBalance(ctx context.Context, userId int) (*model.UserBalanceResponse, error) {
	balance, err := g.GophermartRepo.GetBalance(ctx, userId)
	if err != nil {
		return nil, err
	}
	if balance == nil {
		return nil, BalanceNotFoundError
	}
	return &model.UserBalanceResponse{Current: balance.Current, Withdrawn: balance.Withdrawn}, nil
}

func (g *Service) AddWithdraw(ctx context.Context, userID int, order string, sum float64) error {
	if !utils.OrderNumberCheck(order) {
		return OrderNumberIncorrectError
	}

	balance, tx, err := g.GophermartRepo.GetBalanceTx(ctx, userID)
	if tx != nil {
		defer tx.Rollback()
	}
	if err != nil {
		return err
	}
	if balance == nil {
		return BalanceNotFoundError
	}

	if balance.Current < sum {
		return InsufficientFundsError
	}
	return g.GophermartRepo.AddWithdrawTx(
		ctx,
		tx,
		&repo_model.AddWithdraw{
			BalanceID:    balance.ID,
			UserID:       userID,
			Order:        order,
			Sum:          sum,
			NewCurrent:   balance.Current - sum,
			NewWithdrawn: balance.Withdrawn + sum,
		},
	)
}

func (g *Service) GetWithdrawals(ctx context.Context, userID int) (*[]model.GetWithdrawResponse, error) {
	withdrawals, err := g.GophermartRepo.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}
	var resp []model.GetWithdrawResponse
	for _, w := range *withdrawals {
		resp = append(
			resp,
			model.GetWithdrawResponse{Order: w.Order, Sum: w.Sum, ProcessedAt: w.Processed.Format(time.RFC3339)},
		)
	}
	return &resp, nil
}

func (g *Service) CreateOrder(ctx context.Context, userID int, orderNum string) (bool, error) {
	if orderNum == "" || !utils.OrderNumberCheck(orderNum) {
		return false, OrderNumberIncorrectError
	}

	created, err := g.GophermartRepo.CreateOrder(ctx, userID, orderNum)
	if err != nil {
		return false, err
	}
	if created {
		return true, nil
	}

	existOrder, err := g.GophermartRepo.GetOrderByNumber(ctx, orderNum)
	if err != nil {
		return false, err
	}
	if existOrder.UserId != userID {
		return false, AnotherUserError
	}
	return false, nil
}

func (g *Service) GetOrders(ctx context.Context, userID int) (*[]model.GetOrdersResponse, error) {
	orders, err := g.GophermartRepo.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	var resp []model.GetOrdersResponse
	for _, o := range *orders {
		var accrual float64
		if o.Accrual.Valid {
			accrual = o.Accrual.Float64
		}
		resp = append(
			resp,
			model.GetOrdersResponse{Number: o.Number, Status: string(o.Status), Accrual: accrual, UploadedAt: o.UploadedAt.Format(time.RFC3339)},
		)
	}
	return &resp, nil
}

func NewGophermartService(repo expectedRepo) *Service {
	return &Service{repo}
}
