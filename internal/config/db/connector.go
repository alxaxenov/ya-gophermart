package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type DB interface {
	Querier
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
}

type Tx interface {
	Querier
	Commit() error
	Rollback() error
}

type Connector struct {
	DSN string
	DB  DB
}

func (c *Connector) GetDB() DB {
	return c.DB
}

func (c *Connector) Close() error {
	return c.DB.Close()
}
