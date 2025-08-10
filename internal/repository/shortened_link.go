package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
)

type LinkRepository struct {
	storage map[string]*model.Link
	*config.AppConfig
}

func (r *LinkRepository) Create(link *model.Link) {
	r.storage[link.Shortcut] = link
}

func (r *LinkRepository) GetByShortcut(shortcut string) (*model.Link, bool) {
	l, ok := r.storage[shortcut]
	return l, ok
}

func (r *LinkRepository) LoadStoredData() error {
	var storedData []*model.Link

	file, err := os.OpenFile(r.AppConfig.StoragePath, os.O_RDONLY|os.O_CREATE, 0644)

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

func (r *LinkRepository) SaveInStorage() error {
	var storedData []*model.Link

	file, err := os.OpenFile(r.AppConfig.StoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

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

func NewLinksRepository(config *config.AppConfig) *LinkRepository {
	return &LinkRepository{AppConfig: config, storage: make(map[string]*model.Link)}
}
