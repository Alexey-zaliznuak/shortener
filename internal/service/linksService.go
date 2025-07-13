package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository"
	"gorm.io/gorm"
)

type LinksService struct {
	repository repository.LinkRepository
}

func (s *LinksService) GetFullUrlFromShort(shortUrl string) (string, error) {
	link, err := s.repository.GetByShortUrl(shortUrl)
	if link == nil {
		return "", fmt.Errorf("specified link not found")
	}
	return link.FullUrl, err
}

func (s *LinksService) CreateLink(link *model.Link) error {
	if link.ShortUrl == "" {
		link.ShortUrl = s.generateShortLink(config.Config.ShortLinksLength)
	}

	return s.repository.Create(link)
}

func (s *LinksService) generateShortLink(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.NewSource(time.Now().UnixNano())

	result := make([]rune, length)

	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func NewLinksService(client *gorm.DB) *LinksService {
	return &LinksService{repository: *repository.NewLinkRepository(client)}
}
