package database

import (
	"chat-app/internal/config"
	"database/sql"
)

func GetDB() *sql.DB {
	return config.DB
}
