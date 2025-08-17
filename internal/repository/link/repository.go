package link

import (
	"database/sql"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
)

type LinkRepository interface {
	Create(link *model.Link)
	GetByShortcut(shortcut string) (*model.Link, bool)
	LoadStoredData() error
	SaveInStorage() error
}

func NewLinksRepository(cfg *config.AppConfig, db *sql.DB) LinkRepository {
	if cfg.DB.DatabaseDSN == ""{
		return NewInMemoryLinksRepository(cfg)
	}

	return nil
}
