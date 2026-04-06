package db

import (
	"context"
	"database/sql"
)

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type DB interface {
	Querier
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
}

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks
type Tx interface {
	Querier
	Commit() error
	Rollback() error
}

// Connector коннектор подключения к абстрактой базе данных
type Connector struct {
	DSN string
	DB  DB
}

// GetDB получение сущности, реализующей интерфейс для взаимодействия с бд
func (c *Connector) GetDB() DB {
	return c.DB
}

// Close закрытие соединения с бд
func (c *Connector) Close() error {
	return c.DB.Close()
}
