package link

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"go.uber.org/zap"
)

type PostgreSQLLinksRepository struct {
	db     *sql.DB
	table  string
	config *config.AppConfig
	ctx    context.Context
}

func (r *PostgreSQLLinksRepository) GetByShortcut(shortcut string) (*model.Link, error) {
	result := &model.Link{}

	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(
		ctx,
		fmt.Sprintf(
			`
			SELECT url, shortcut
			FROM %s
			WHERE shortcut = $1
			`,
			r.table,
		),
		shortcut,
	)

	err := row.Scan(&result.FullURL, &result.Shortcut)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound // или своя доменная ошибка NotFound
		}
		return nil, err
	}
	return result, nil
}

func (r *PostgreSQLLinksRepository) getAll() ([]*model.Link, error) {
	result := []*model.Link{}

	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(
		ctx,
		fmt.Sprintf(
			`
			SELECT url, shortcut
			FROM %s
			`,
			r.table,
		),
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		l := &model.Link{}
		err = rows.Scan(&l.FullURL, &l.Shortcut)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("row scanning failing: %s", err.Error()))
			continue
		}

		result = append(result, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, err
}

func (r *PostgreSQLLinksRepository) Create(link *model.Link) error {
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (url, shortcut) VALUES ($1, $2)", r.table), link.FullURL, link.Shortcut)
	return err
}

func (r *PostgreSQLLinksRepository) LoadStoredData() error {
	var storedData []*model.Link
	var restored, skipped int

	file, err := os.OpenFile(r.config.DB.StoragePath, os.O_RDONLY|os.O_CREATE, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&storedData)

	if err != nil {
		if errors.Is(err, io.EOF) {
			logger.Log.Warn("Empty file storage")
		} else {
			return err
		}
	}

	for _, link := range storedData {
		_, err := r.GetByShortcut(link.Shortcut)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				err := r.Create(link)
				if err != nil {
					logger.Log.Error(fmt.Sprintf("Failing link creation: %s", err.Error()))
					skipped += 1
					continue
				}
				restored += 1
				continue
			}
			logger.Log.Error(fmt.Sprintf("Failing link creation: %s", err.Error()))
		}
		skipped += 1
	}

	logger.Log.Info("Restored urls", zap.Int("restored", restored), zap.Int("skipped", skipped))

	return nil
}

func (r *PostgreSQLLinksRepository) SaveInStorage() error {
	file, err := os.OpenFile(r.config.DB.StoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	allLinks, err := r.getAll()

	if err != nil {
		return err
	}

	err = json.NewEncoder(file).Encode(&allLinks)

	if err != nil {
		return err
	}

	logger.Log.Info(fmt.Sprintf("Saved urls: %d", len(allLinks)))

	return nil
}

func NewInPostgresSQLLinksRepository(ctx context.Context, config *config.AppConfig, db *sql.DB) (*PostgreSQLLinksRepository, error) {
	return &PostgreSQLLinksRepository{
		db:     db,
		config: config,
		table:  "links",
		ctx:    ctx,
	}, nil
}
