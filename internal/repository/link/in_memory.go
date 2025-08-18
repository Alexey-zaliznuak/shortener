package link

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
)

type InMemoryLinkRepository struct {
	storage map[string]*model.Link
	mu      sync.RWMutex
	config  *config.AppConfig
}

func (r *InMemoryLinkRepository) Create(link *model.Link) error {
	r.mu.Lock()
	r.storage[link.Shortcut] = link
	r.mu.Unlock()
	return nil
}

func (r *InMemoryLinkRepository) GetByShortcut(shortcut string) (*model.Link, error) {
	r.mu.RLock()
	l, ok := r.storage[shortcut]
	r.mu.RUnlock()

	if ok {
		return l, nil
	}
	return l, database.ErrNotFound
}

func (r *InMemoryLinkRepository) LoadStoredData() error {
	var storedData []*model.Link

	file, err := os.OpenFile(r.config.DB.StoragePath, os.O_RDONLY|os.O_CREATE, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&storedData)

	if err != nil {
		return err
	}

	for _, link := range storedData {
		r.Create(link)
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

	defer file.Close()

	for _, link := range r.storage {
		storedData = append(storedData, link)
	}

	err = json.NewEncoder(file).Encode(&storedData)

	if err != nil {
		return err
	}

	logger.Log.Info(fmt.Sprintf("Saved urls: %d", len(storedData)))

	return nil
}

func NewInMemoryLinksRepository(config *config.AppConfig) *InMemoryLinkRepository {
	return &InMemoryLinkRepository{config: config, storage: make(map[string]*model.Link)}
}
