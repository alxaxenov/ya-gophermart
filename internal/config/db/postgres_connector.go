package db

import (
	"fmt"

	"github.com/alxaxenov/ya-gophermart/migrations"
	"github.com/jmoiron/sqlx"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ConnectorPG struct {
	Connector
}

func (c *ConnectorPG) Open() (*sqlx.DB, error) {
	dataBase, err := sqlx.Connect("pgx", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	c.DB = &PGDBWrapper{dataBase}
	if err := c.Migrate(dataBase); err != nil {
		defer dataBase.Close()
		return nil, err
	}
	return dataBase, nil
}

func (c *ConnectorPG) Migrate(dataBase *sqlx.DB) error {
	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migration SetDialect error: %w", err)
	}
	if err := goose.Up(dataBase.DB, "."); err != nil {
		return fmt.Errorf("migration Up error: %w", err)
	}
	return nil
}

func NewPGConnector(dsn string) *ConnectorPG {
	return &ConnectorPG{Connector: Connector{DSN: dsn}}
}
