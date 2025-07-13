package service

import (
	"fmt"
	"math/rand"
	"net/url"
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
	if !s.isValidURL(link.FullUrl) {
		return fmt.Errorf("create link error: invalid url: '%s'", link.FullUrl)
	}

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

func (s *LinksService) isValidURL(u string) bool {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	return true
}

func NewLinksService(client *gorm.DB) *LinksService {
	return &LinksService{repository: *repository.NewLinkRepository(client)}
}
