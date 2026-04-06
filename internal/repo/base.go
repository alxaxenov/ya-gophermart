package repo

import "github.com/alxaxenov/ya-gophermart/internal/config/db"

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks
type connector interface {
	GetDB() db.DB
}
