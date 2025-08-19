package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	iofs "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrNotFound = errors.New("not found")

func NewDatabaseConnectionPool(cfg *config.AppConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DB.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate driver: %w", err)
	}

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate new: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		_ = db.Close()
		return nil, fmt.Errorf("migrate up: %w", err)
	}

	if v, dirty, err := m.Version(); err != nil {
		logger.Log.Info(fmt.Sprintf("get migration version error: %s", err.Error()))
	} else {
		logger.Log.Info(fmt.Sprintf("migration info: version=%d dirty=%v\n", v, dirty))
	}

	return db, nil
}
