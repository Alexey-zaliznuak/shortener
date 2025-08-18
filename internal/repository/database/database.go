package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrNotFound = errors.New("not found")

func NewDatabaseConnectionPool(cfg *config.AppConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DB.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}
