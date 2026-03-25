package model

import (
	"database/sql"
	"time"

	"github.com/alxaxenov/ya-gophermart/internal/domain/order"
)

type Balance struct {
	ID        int     `db:"id"`
	UserId    int     `db:"user_id"`
	Current   float64 `db:"current"`
	Withdrawn float64 `db:"withdrawn"`
}

type AddWithdraw struct {
	BalanceID    int
	UserID       int
	Order        string
	Sum          float64
	NewCurrent   float64
	NewWithdrawn float64
}

type Withdraw struct {
	Order     string    `db:"order_number"`
	Sum       float64   `db:"sum"`
	Processed time.Time `db:"processed_at"`
}

type Order struct {
	Id     int `db:"id"`
	UserId int `db:"user_id"`
}

type OrderForList struct {
	Number     string          `db:"number"`
	Status     order.Status    `db:"status"`
	Accrual    sql.NullFloat64 `db:"accrual"`
	UploadedAt time.Time       `db:"uploaded_at"`
}

type OrderToProcess struct {
	Number string       `db:"number"`
	Status order.Status `db:"status"`
	UserID int          `db:"user_id"`
}
