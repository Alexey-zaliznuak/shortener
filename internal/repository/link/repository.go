package link

import (
	"context"
	"database/sql"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
)

type LinkRepository interface {
	GetByShortcut(shortcut string) (*model.Link, error)
	Create(link *model.Link, executer database.Executer) (*model.Link, bool, error)
	LoadStoredData() error
	SaveInStorage() error
	GetTransactionExecuter(ctx context.Context, opts *sql.TxOptions) (database.TransactionExecuter, error)
}

func NewLinksRepository(ctx context.Context, cfg *config.AppConfig, db *sql.DB) (LinkRepository, error) {
	if cfg.DB.DatabaseDSN == "" {
		return NewInMemoryLinksRepository(cfg), nil
	}

	return NewInPostgresSQLLinksRepository(ctx, cfg, db)
}
