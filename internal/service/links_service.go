package service

import (
	"context"
	"errors"
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
	auth *AuthService
	*config.AppConfig
}

func (s *LinksService) GetFullURLFromShort(shortcut string) (string, error) {
	link, err := s.repository.GetByShortcut(shortcut)
	if err != nil {
		return "", err
	}
	return link.FullURL, nil
}

func (s *LinksService) GetUserLinks(c *gin.Context) ([]*model.GetUserLinksRequestItem, error) {
	claims, err := s.auth.GetAuthorization(c)

	if err != nil {
		return nil, err
	}

	return s.repository.GetByUserID(claims.UserID)
}

func (s *LinksService) CreateLink(link *model.CreateLinkDto, c *gin.Context) (*model.Link, bool, error) {
	auth, err := s.auth.GetOrCreateAndSaveAuthorization(c)

	if err != nil {
		return nil, false, err
	}

	if !s.isValidURL(link.FullURL) {
		return link.NewLink(auth.UserID), false, fmt.Errorf("create link error: invalid URL: '%s'", link.FullURL)
	}

	if link.Shortcut == "" {
		var err error

		link.Shortcut, err = s.createUniqueShortcut()

		if err != nil {
			return link.NewLink(auth.UserID), false, err
		}
	}

	return s.repository.Create(link, auth.UserID, nil)
}

func (s *LinksService) BulkCreateWithCorrelationID(links []*model.CreateLinkWithCorrelationIDRequestItem, c *gin.Context) ([]*model.CreateLinkWithCorrelationIDResponseItem, error) {
	var result []*model.CreateLinkWithCorrelationIDResponseItem

	auth, err := s.auth.GetOrCreateAndSaveAuthorization(c)

	if err != nil {
		return nil, err
	}

	transactionExecuter, err := s.repository.GetTransactionExecuter(context.Background(), nil)
	supportTransaction := true

	if err != nil {
		if errors.Is(err, database.ErrExecuterNotSupportTransactions) {
			supportTransaction = false
		} else {
			return nil, err
		}
	}

	for index, link := range links {
		if !s.isValidURL(link.FullURL) {
			if supportTransaction {
				transactionExecuter.Commit()
			}
			return nil, fmt.Errorf("create link error: invalid URL: '%s'", link.FullURL)
		}

		shortcut, err := s.createUniqueShortcut()
		if err != nil {
			if supportTransaction {
				transactionExecuter.Commit()
			}
			return nil, err
		}

		l := &model.CreateLinkDto{FullURL: link.FullURL, Shortcut: shortcut}

		newLink, _, err := s.repository.Create(l, auth.UserID, transactionExecuter)

		if err != nil {
			if supportTransaction {
				transactionExecuter.Commit()
			}
			return nil, err
		}

		shortcut, err = s.BuildShortURL(newLink.Shortcut, c)

		if err != nil {
			if supportTransaction {
				transactionExecuter.Commit()
			}
			return nil, err
		}

		result = append(result, &model.CreateLinkWithCorrelationIDResponseItem{CorrelationID: link.CorrelationID, Shortcut: shortcut})

		if supportTransaction && (index+1%1000 == 0 || index == len(links)-1) {
			transactionExecuter.Commit()
			transactionExecuter, err = s.repository.GetTransactionExecuter(context.Background(), nil)
			if err != nil {
				return nil, err
			}
		}
	}

	if supportTransaction {
		transactionExecuter.Commit()
	}

	return result, nil
}

func (s *LinksService) createUniqueShortcut() (string, error) {
	maxAttempts := 5

	newShortcut := s.generateShortcut(s.AppConfig.Server.ShortLinksLength)

	for range maxAttempts {
		_, err := s.repository.GetByShortcut(newShortcut)
		if err != nil {
			newShortcut = s.generateShortcut(s.AppConfig.Server.ShortLinksLength)
			continue
		}
		break
	}

	if _, err := s.repository.GetByShortcut(newShortcut); err != database.ErrNotFound {
		logger.Log.Error(err.Error())
		return "", fmt.Errorf("create link error: could not generate unique shortcut after %d attempts", maxAttempts)
	}

	return newShortcut, nil
}

func (s *LinksService) generateShortcut(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	result := make([]rune, length)

	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func (s *LinksService) BuildShortURL(shortcut string, c *gin.Context) (string, error) {
	prefix := s.AppConfig.Server.BaseURL
	if prefix == "" {
		prefix = fmt.Sprintf("http://%s/", c.Request.Host)
	}
	return url.JoinPath(prefix, shortcut)
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
		auth: NewAuthService(config),
		AppConfig:  config,
	}
}
