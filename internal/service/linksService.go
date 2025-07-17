package service

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LinksService struct {
	repository repository.LinkRepository
}

func (s *LinksService) GetFullURLFromShort(shortcut string) (string, error) {
	link, err := s.repository.GetByShortcut(shortcut)
	if link == nil {
		return "", fmt.Errorf("specified link not found")
	}
	return link.FullURL, err
}

func (s *LinksService) CreateLink(link *model.Link) error {
	if !s.isValidURL(link.FullURL) {
		return fmt.Errorf("create link error: invalid URL: '%s'", link.FullURL)
	}

	if link.Shortcut == "" {
		link.Shortcut = s.generateShortcut(config.GetConfig().ShortLinksLength)
	}

	flag := true
	for flag {
		_, err := s.repository.GetByShortcut(link.Shortcut)
		if err == nil {
			link.Shortcut = s.generateShortcut(config.GetConfig().ShortLinksLength)
			continue
		}
		flag = false
	}

	return s.repository.Create(link)
}

func (s *LinksService) BuildShortURL(shortcut string, c *gin.Context) (string, error) {
	base := config.GetConfig().ShortLinksURLPrefix
	if base == "" {
		base = fmt.Sprintf("http://%s/", c.Request.Host)
	}

	return url.JoinPath(base, shortcut)
}

func (s *LinksService) generateShortcut(length int) string {
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
