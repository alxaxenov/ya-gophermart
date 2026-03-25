package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/alxaxenov/ya-gophermart/internal/config/db"
	"github.com/alxaxenov/ya-gophermart/internal/domain/order"
	"github.com/alxaxenov/ya-gophermart/internal/repo/model"
)

type GophermartRepo struct {
	Connector connector
}

func (g *GophermartRepo) GetBalance(ctx context.Context, userID int) (*model.Balance, error) {
	db := g.Connector.GetDB()
	var balance model.Balance
	query := "SELECT id, user_id, current, withdrawn FROM balance WHERE user_id = $1"
	err := db.GetContext(ctx, &balance, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("GophermartRepo GetBalance select error: %w", err)
	}
	return &balance, nil
}

func (g *GophermartRepo) GetBalanceTx(ctx context.Context, userID int) (*model.Balance, db.Tx, error) {
	tx, err := g.Connector.GetDB().BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, nil, fmt.Errorf("GophermartRepo GetBalanceTx begin tx error: %w", err)
	}

	var balance model.Balance
	query := "SELECT id, user_id, current, withdrawn FROM balance WHERE user_id = $1 FOR UPDATE"
	err = tx.GetContext(ctx, &balance, query, userID)
	if err != nil {
		defer tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("GophermartRepo GetBalanceTx select error: %w", err)
	}
	return &balance, tx, nil
}

func (g *GophermartRepo) AddWithdrawTx(ctx context.Context, tx db.Tx, withdraw *model.AddWithdraw) error {
	defer tx.Rollback()

	queryWithdraw := "INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)"
	queryBalance := "UPDATE balance SET current = $1, withdrawn = $2 WHERE id = $3"

	_, err := tx.ExecContext(ctx, queryWithdraw, withdraw.UserID, withdraw.Order, withdraw.Sum)
	if err != nil {
		return fmt.Errorf("GophermartRepo AddWithdrawTx withdrawals query error: %w", err)
	}

	_, err = tx.ExecContext(ctx, queryBalance, withdraw.NewCurrent, withdraw.NewWithdrawn, withdraw.BalanceID)
	if err != nil {
		return fmt.Errorf("GophermartRepo AddWithdrawTx withdraw query error: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("GophermartRepo AddWithdrawTx tx commit error: %w", err)
	}
	return nil
}

func (g *GophermartRepo) GetUserWithdrawals(ctx context.Context, userID int) (*[]model.Withdraw, error) {
	db := g.Connector.GetDB()
	var withdraws []model.Withdraw
	query := "SELECT order_number, sum, processed_at from withdrawals WHERE user_id = $1 ORDER BY processed_at DESC"
	err := db.SelectContext(ctx, &withdraws, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GophermartRepo GetUserWithdrawals select error: %w", err)
	}
	return &withdraws, nil
}

func (g *GophermartRepo) CreateOrder(ctx context.Context, userID int, orderNum string) (bool, error) {
	db := g.Connector.GetDB()
	query := "INSERT INTO orders (user_id, number) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id"
	raw, err := db.ExecContext(ctx, query, userID, orderNum)
	if err != nil {
		return false, fmt.Errorf("GophermartRepo CreateOrder exec error: %w", err)
	}
	affected, err := raw.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("GophermartRepo CreateOrder rows affected error: %w", err)
	}
	if affected == 0 {
		return false, nil
	}
	return true, nil
}

func (g *GophermartRepo) GetOrderByNumber(ctx context.Context, orderNum string) (*model.Order, error) {
	db := g.Connector.GetDB()
	var order model.Order
	query := "SELECT id, user_id FROM orders WHERE number = $1"
	err := db.GetContext(ctx, &order, query, orderNum)
	if err != nil {
		return nil, fmt.Errorf("GophermartRepo GetOrderByNumber select error: %w", err)
	}
	return &order, nil
}

func (g *GophermartRepo) GetOrders(ctx context.Context, userID int) (*[]model.OrderForList, error) {
	db := g.Connector.GetDB()
	var orders []model.OrderForList
	query := "SELECT number, status, accrual, uploaded_at from orders WHERE user_id = $1 ORDER BY uploaded_at DESC"
	err := db.SelectContext(ctx, &orders, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GophermartRepo GetOrders select error: %w", err)
	}
	return &orders, nil
}

func (g *GophermartRepo) GetOrdersToProcess(ctx context.Context, limit int) (*[]model.OrderToProcess, error) {
	db := g.Connector.GetDB()
	var orders []model.OrderToProcess
	query := "SELECT number, status, user_id FROM orders WHERE status IN ($1, $2) ORDER BY uploaded_at ASC LIMIT $3"
	err := db.SelectContext(ctx, &orders, query, order.NEW, order.PROCESSING, limit)
	if err != nil {
		return nil, fmt.Errorf("GophermartRepo GetOrdersToProcess select error: %w", err)
	}
	return &orders, nil
}

func (g *GophermartRepo) UpdateOrderStatus(ctx context.Context, orderNum string, status order.Status) error {
	db := g.Connector.GetDB()
	query := "UPDATE orders SET status = $1, updated_at = $2 WHERE number = $3"
	_, err := db.ExecContext(ctx, query, status, time.Now(), orderNum)
	if err != nil {
		return fmt.Errorf("GophermartRepo UpdateOrderStatus update error: %w", err)
	}
	return nil
}

func (g *GophermartRepo) UpdateOrderAccrual(ctx context.Context, orderNum string, userID int, status order.Status, accrual float64) error {
	tx, err := g.Connector.GetDB().BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("GophermartRepo UpdateOrderAccrual begin tx error: %w", err)
	}

	queryOrder := "UPDATE orders SET accrual = $1, status = $2, updated_at = $3 WHERE number = $4"
	queryBalance := "UPDATE balance SET current = current + $1 WHERE user_id = $2"

	_, err = tx.ExecContext(ctx, queryOrder, accrual, status, time.Now(), orderNum)
	if err != nil {
		return fmt.Errorf("GophermartRepo UpdateOrderAccrual queryOrder error: %w", err)
	}
	_, err = tx.ExecContext(ctx, queryBalance, accrual, userID)
	if err != nil {
		return fmt.Errorf("GophermartRepo UpdateOrderAccrual queryBalance error: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("GophermartRepo UpdateOrderAccrual tx commit error: %w", err)
	}
	return nil
}

func NewGophermartRepo(connector connector) *GophermartRepo {
	return &GophermartRepo{connector}
}
