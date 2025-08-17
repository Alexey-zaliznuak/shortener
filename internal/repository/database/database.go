package database

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Executer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewDatabaseConnectionPool(cfg *config.AppConfig) *sql.DB {
	db, err := sql.Open("pgx", cfg.DB.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Не удалось подключиться:", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	return db
}
