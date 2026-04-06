package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alxaxenov/ya-gophermart/internal/repo/model"
)

// UserRepo репозиторий для взаимодействия с таблицей users
type UserRepo struct {
	Connector connector
}

// CheckLogin проверка существует ли уже пользователь с таким логином
func (d *UserRepo) CheckLogin(ctx context.Context, login string) (bool, error) {
	db := d.Connector.GetDB()
	var userID int
	err := db.GetContext(ctx, &userID, "SELECT id FROM users WHERE login = $1", login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("UserRepo CheckLogin select error: %w", err)
	}
	return userID != 0, nil
}

// CreateUser создание нового пользователя и его баланса
func (d *UserRepo) CreateUser(ctx context.Context, login string, password string) (int, error) {
	db := d.Connector.GetDB()
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		return 0, fmt.Errorf("UserRepo CreateUser begin tx error: %w", err)
	}
	defer tx.Rollback()

	queryUser := "INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id"
	queryBalance := "INSERT INTO balance (user_id) VALUES ($1)"

	var userID int
	err = tx.GetContext(ctx, &userID, queryUser, login, password)
	if err != nil {
		return 0, fmt.Errorf("UserRepo CreateUser user query error: %w", err)
	}

	_, err = tx.ExecContext(ctx, queryBalance, userID)
	if err != nil {
		return 0, fmt.Errorf("UserRepo CreateUser balance query error: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("UserRepo CreateUser commit error: %w", err)
	}
	return userID, nil
}

// GetUser получение существующего пользователя по его логину
func (d *UserRepo) GetUser(ctx context.Context, login string) (*model.User, error) {
	db := d.Connector.GetDB()
	var user model.User
	query := "SELECT id, login, password_hash, active FROM users WHERE login = $1"
	err := db.GetContext(ctx, &user, query, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepo GetUser select error: %w", err)
	}
	return &user, nil
}

func NewUserRepo(connector connector) *UserRepo {
	return &UserRepo{connector}
}
