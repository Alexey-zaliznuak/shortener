package service

import (
	"fmt"
	"math/rand"
	"net/url"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/link"
	"github.com/gin-gonic/gin"
)

type LinksService struct {
	repository link.LinkRepository
	*config.AppConfig
}

func (s *LinksService) GetFullURLFromShort(shortcut string) (string, error) {
	link, err := s.repository.GetByShortcut(shortcut)
	if err != nil {
		return "", err
	}
	return link.FullURL, nil
}

func (s *LinksService) CreateLink(link *model.Link) error {
	if !s.isValidURL(link.FullURL) {
		return fmt.Errorf("create link error: invalid URL: '%s'", link.FullURL)
	}

	if link.Shortcut == "" {
		link.Shortcut = s.generateShortcut(s.AppConfig.Server.ShortLinksLength)
	}

	const maxAttempts = 5
	for range maxAttempts {
		_, err := s.repository.GetByShortcut(link.Shortcut)
		if err != nil {
			link.Shortcut = s.generateShortcut(s.AppConfig.Server.ShortLinksLength)
			continue
		}
		break
	}

	if _, err := s.repository.GetByShortcut(link.Shortcut); err != database.ErrNotFound {
		logger.Log.Error(err.Error())
		return fmt.Errorf("create link error: could not generate unique shortcut after %d attempts", maxAttempts)
	}

	s.repository.Create(link)
	return nil
}

func (s *LinksService) BuildShortURL(shortcut string, c *gin.Context) (string, error) {
	prefix := s.AppConfig.Server.BaseURL
	if prefix == "" {
		prefix = fmt.Sprintf("http://%s/", c.Request.Host)
	}
	return url.JoinPath(prefix, shortcut)
}

func (s *LinksService) generateShortcut(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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

func NewLinksService(repository link.LinkRepository, config *config.AppConfig) *LinksService {
	return &LinksService{
		repository: repository,
		AppConfig:  config,
	}
}
