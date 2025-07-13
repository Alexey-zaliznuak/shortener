package service

import (
	"fmt"

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

func NewLinksService(client *gorm.DB) *LinksService {
	return &LinksService{repository: *repository.NewLinkRepository(client)}
}
