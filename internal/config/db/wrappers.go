package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type PGDBWrapper struct {
	*sqlx.DB
}

func (w *PGDBWrapper) BeginTxx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := w.DB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
