package link

import (
	"database/sql"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
)

type PostgreSQLLinksRepository struct {
	db     *sql.DB
	table  string
	config *config.AppConfig
}

func (r *PostgreSQLLinksRepository) Create(link *model.Link) {
	// r.db.
}

func (r *PostgreSQLLinksRepository) GetByShortcut(shortcut string) (*model.Link, bool) {
	return nil, true
}

func (r *PostgreSQLLinksRepository) LoadStoredData() error {
	// var storedData []*model.Link

	// file, err := os.OpenFile(r.config.DB.StoragePath, os.O_RDONLY|os.O_CREATE, 0644)

	// if err != nil {
	// 	return err
	// }

	// defer file.Close()

	// err = json.NewDecoder(file).Decode(&storedData)

	// if err != nil {
	// 	return err
	// }

	// for _, link := range storedData {
	// 	r.Create(link)
	// }

	// logger.Log.Info(fmt.Sprintf("Restored urls: %d", len(storedData)))

	return nil
}

func (r *PostgreSQLLinksRepository) SaveInStorage() error {
	// var storedData []*model.Link

	// file, err := os.OpenFile(r.config.DB.StoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	// if err != nil {
	// 	return err
	// }

	// defer file.Close()

	// for _, link := range r.config {
	// 	storedData = append(storedData, link)
	// }

	// err = json.NewEncoder(file).Encode(&storedData)

	// if err != nil {
	// 	return err
	// }

	// logger.Log.Info(fmt.Sprintf("Saved urls: %d", len(storedData)))

	return nil
}

func NewInPostgresSQLLinksRepository(db *sql.DB, config *config.AppConfig) *PostgreSQLLinksRepository {
	db.Exec(`CREATE TABLE IF NOT EXISTS Links (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), "url" TEXT, "shortcut" TEXT)`)
	return &PostgreSQLLinksRepository{
		db:     db,
		config: config,
		table:  `"Link"`,
	}
}
