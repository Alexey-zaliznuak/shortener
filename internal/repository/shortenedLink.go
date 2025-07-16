package repository

import (
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"gorm.io/gorm"
)

type LinkRepository struct {
	Client *gorm.DB
}

func (r *LinkRepository) Create(link *model.Link) error {
	return r.Client.Create(link).Error
}

func (r *LinkRepository) GetByShortURL(shortURL string) (*model.Link, error) {
	link := &model.Link{}
	err := r.Client.Where("short_URL = ?", shortURL).First(link).Error

	if err != nil {
		return nil, err
	}

	return link, nil
}

func NewLinkRepository(client *gorm.DB) *LinkRepository {
	return &LinkRepository{Client: client}
}
