package link

import (
	"context"
	"database/sql"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
)

type LinkRepository interface {
	Create(link *model.Link) error
	GetByShortcut(shortcut string) (*model.Link, error)
	LoadStoredData() error
	SaveInStorage() error
}

func NewLinksRepository(ctx context.Context, cfg *config.AppConfig, db *sql.DB) (LinkRepository, error) {
	if cfg.DB.DatabaseDSN == "" {
		return NewInMemoryLinksRepository(cfg), nil
	}

	return NewInPostgresSQLLinksRepository(ctx, cfg, db)
}
