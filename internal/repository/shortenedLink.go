package repository

import (
	"errors"

	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"gorm.io/gorm"
)

type LinkRepository struct {
	Client *gorm.DB
}

func (r *LinkRepository) Create(link *model.Link) error {
	return r.Client.Create(link).Error
}

func (r *LinkRepository) GetByShortUrl(shortUrl string) (*model.Link, error) {
	link := &model.Link{}
	err := r.Client.Where("short_url = ?", shortUrl).First(link).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return link, nil
}

func NewLinkRepository(client *gorm.DB) *LinkRepository {
	return &LinkRepository{Client: client}
}
