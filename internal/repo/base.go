package repo

import "github.com/alxaxenov/ya-gophermart/internal/config/db"

type connector interface {
	GetDB() db.DB
}
