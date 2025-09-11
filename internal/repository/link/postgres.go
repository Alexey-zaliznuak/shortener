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
	"github.com/Alexey-zaliznuak/shortener/internal/utils"
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

func (r *PostgreSQLLinksRepository) GetByUserID(userID string) ([]*model.GetUserLinksRequestItem, error) {
	result := []*model.GetUserLinksRequestItem{}

	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(
		ctx,
		fmt.Sprintf(
			`
			SELECT url, shortcut
			FROM %s
			WHERE userID = $1
			`,
			r.table,
		),
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer func() { utils.LogErrorWrapper(rows.Close()) }()

	for rows.Next() {
		l := &model.GetUserLinksRequestItem{}
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

func (r *PostgreSQLLinksRepository) getAll() ([]*model.Link, error) {
	result := []*model.Link{}

	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(
		ctx,
		fmt.Sprintf(
			`
			SELECT url, shortcut, userID
			FROM %s
			`,
			r.table,
		),
	)

	if err != nil {
		return nil, err
	}

	defer func() { utils.LogErrorWrapper(rows.Close()) }()

	for rows.Next() {
		l := &model.Link{}
		err = rows.Scan(&l.FullURL, &l.Shortcut, &l.UserID)
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

// Will modify shortcut if find link with same full url
func (r *PostgreSQLLinksRepository) Create(link *model.CreateLinkDto, UserID string, executer database.Executer) (*model.Link, bool, error) {
	var exec database.Executer = r.db
	oldShortcut := link.Shortcut

	if executer != nil {
		exec = executer
	}

	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	// TODO: with precompiled queries
	res, err := r.QueryRowContextWithRetry(ctx, fmt.Sprintf(
		`INSERT INTO %s (url, shortcut, userID)
			VALUES ($1, $2, $3)
			ON CONFLICT (url) DO UPDATE SET shortcut = links.shortcut
			RETURNING %s.url, %s.shortcut, %s.userID;
			`,
		r.table, r.table, r.table, r.table,
	),
		exec,
		link.FullURL,
		link.Shortcut,
		UserID,
	)

	if err != nil {
		return link.NewLink(UserID), false, err
	}

	newLink := &model.Link{}
	res.Scan(&newLink.FullURL, &newLink.Shortcut, &newLink.UserID)

	return newLink, oldShortcut == newLink.Shortcut, nil
}

func (r *PostgreSQLLinksRepository) LoadStoredData() error {
	var storedData []*model.Link
	var restored, skipped int

	file, err := os.OpenFile(r.config.DB.StoragePath, os.O_RDONLY|os.O_CREATE, 0644)

	if err != nil {
		return err
	}

	defer func() { utils.LogErrorWrapper(file.Close()) }()

	err = json.NewDecoder(file).Decode(&storedData)

	if err != nil {
		if errors.Is(err, io.EOF) {
			logger.Log.Warn("Empty file storage")
		} else {
			return err
		}
	}

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	for _, link := range storedData {
		_, err := r.GetByShortcut(link.Shortcut)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				res, err := r.QueryRowContextWithRetry(
					context.Background(),
					fmt.Sprintf(
						`INSERT INTO %s (url, shortcut, userID)
						VALUES ($1, $2, $3)
						ON CONFLICT (url) DO NOTHING
						RETURNING %s.url, %s.shortcut;
					`, r.table, r.table, r.table),
					tx,
					link.FullURL,
					link.Shortcut,
					link.UserID,
				)

				if err != nil {
					logger.Log.Error(fmt.Sprintf("Failing link creation: %s", err.Error()))
					tx.Rollback()
					return err
				}

				res.Scan(&link.FullURL, &link.Shortcut)

				restored += 1
				continue
			}
			logger.Log.Error(fmt.Sprintf("Failing link creation: %s", err.Error()))
		}
		skipped += 1
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	logger.Log.Info("Restored urls", zap.Int("restored", restored), zap.Int("skipped", skipped))

	return nil
}

func (r *PostgreSQLLinksRepository) SaveInStorage() error {
	file, err := os.OpenFile(r.config.DB.StoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	defer func() { utils.LogErrorWrapper(file.Close()) }()

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

func (r *PostgreSQLLinksRepository) QueryRowContextWithRetry(ctx context.Context, query string, executer database.Executer, args ...any) (*sql.Row, error) {
	const maxRetries = 3
	var lastErr error

	classifier := database.NewPostgresErrorClassifier()

	for range maxRetries {
		row := executer.QueryRowContext(ctx, query, args...)
		err := row.Err()

		if err == nil {
			return row, nil
		}

		classification := classifier.Classify(err)

		if classification == database.NonRetriable {
			fmt.Printf("Непредвиденная ошибка: %v\n", err)
			return nil, err
		}
	}

	return nil, fmt.Errorf("операция прервана после %d попыток: %w", maxRetries, lastErr)
}

func (r *PostgreSQLLinksRepository) GetTransactionExecuter(ctx context.Context, opts *sql.TxOptions) (database.TransactionExecuter, error) {
	return r.db.BeginTx(ctx, opts)
}

func NewInPostgresSQLLinksRepository(ctx context.Context, config *config.AppConfig, db *sql.DB) (*PostgreSQLLinksRepository, error) {
	return &PostgreSQLLinksRepository{
		db:     db,
		config: config,
		table:  "links",
		ctx:    ctx,
	}, nil
}
