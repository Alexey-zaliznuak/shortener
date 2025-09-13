package link

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/utils"
)

type InMemoryLinkRepository struct {
	shortStorage map[string]*model.Link
	fullStorage  map[string]*model.Link

	shortMu sync.RWMutex
	fullMu  sync.RWMutex

	config *config.AppConfig
}

func (r *InMemoryLinkRepository) GetByShortcut(shortcut string) (*model.Link, error) {
	r.shortMu.RLock()
	l, ok := r.shortStorage[shortcut]
	r.shortMu.RUnlock()

	if ok {
		if l.IsDeleted {
			return nil, database.ErrObjectDeleted
		}
		return l, nil
	}
	return l, database.ErrNotFound
}

func (r *InMemoryLinkRepository) GetByUserID(userID string) ([]*model.GetUserLinksRequestItem, error) {
	var result []*model.GetUserLinksRequestItem

	r.shortMu.RLock()
	for _, val := range r.shortStorage {
		if val.UserID == userID {
			result = append(result, &model.GetUserLinksRequestItem{Shortcut: val.Shortcut, FullURL: val.FullURL})
		}
	}
	r.shortMu.RUnlock()

	return result, nil
}

func (r *InMemoryLinkRepository) GetByFullURL(url string) (*model.Link, error) {
	r.fullMu.RLock()
	l, ok := r.fullStorage[url]
	r.fullMu.RUnlock()

	if ok {
		return l, nil
	}
	return l, database.ErrNotFound
}

func (r *InMemoryLinkRepository) Create(link *model.CreateLinkDto, UserID string, executer database.Executer) (*model.Link, bool, error) {
	l, err := r.GetByFullURL(link.FullURL)

	if err != database.ErrNotFound {
		if err != nil {
			return link.NewLink(UserID), false, err
		}

		return l, false, err
	}

	newLink := link.NewLink(UserID)

	r.shortMu.Lock()
	r.shortStorage[link.Shortcut] = newLink
	r.shortMu.Unlock()

	r.fullMu.Lock()
	r.fullStorage[link.FullURL] = newLink
	r.fullMu.Unlock()

	return newLink, true, nil
}

func (r *InMemoryLinkRepository) DeleteUserLinks(shortcuts []string, userID string) error {
	for _, shortcut := range shortcuts {
		r.shortMu.RLock()

		link, ok := r.shortStorage[shortcut]

		if !ok {
			return database.ErrNotFound
		}

		if link.UserID == userID {
			link.IsDeleted = true
		}

		r.shortMu.RUnlock()
	}

	return nil
}

func (r *InMemoryLinkRepository) LoadStoredData() error {
	var storedData []*model.Link

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

	for _, link := range storedData {
		r.Create(link.ToCreateDto(), "", nil)
	}

	logger.Log.Info(fmt.Sprintf("Restored urls: %d", len(storedData)))

	return nil
}

func (r *InMemoryLinkRepository) SaveInStorage() error {
	var storedData []*model.Link

	file, err := os.OpenFile(r.config.DB.StoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	defer func() { utils.LogErrorWrapper(file.Close()) }()

	for _, link := range r.shortStorage {
		storedData = append(storedData, link)
	}

	err = json.NewEncoder(file).Encode(&storedData)

	if err != nil {
		return err
	}

	logger.Log.Info(fmt.Sprintf("Saved urls: %d", len(storedData)))

	return nil
}

func (r *InMemoryLinkRepository) GetTransactionExecuter(ctx context.Context, opts *sql.TxOptions) (database.TransactionExecuter, error) {
	return nil, database.ErrExecuterNotSupportTransactions
}

func NewInMemoryLinksRepository(config *config.AppConfig) *InMemoryLinkRepository {
	return &InMemoryLinkRepository{config: config, shortStorage: make(map[string]*model.Link), fullStorage: make(map[string]*model.Link)}
}
