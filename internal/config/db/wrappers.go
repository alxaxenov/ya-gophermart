package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// PGDBWrapper обертка для ConnectorPG для возможности использования единого интерфейса запросов Querier при использовании транзакции
type PGDBWrapper struct {
	*sqlx.DB
}

// BeginTxx инициализация транзакции
func (w *PGDBWrapper) BeginTxx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := w.DB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
