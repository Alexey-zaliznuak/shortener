package service

import (
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthService struct {
	Repository *repository.AuthRepository
}

func (service *AuthService) GetAuthorization(c *gin.Context) (*repository.Claims, error) {
	auth, err := c.Cookie("Authorization")

	logger.Log.Info("COOK AUTH:", zap.String("auth", auth))

	if err != nil {
		if err != http.ErrNoCookie {
			return nil, err
		}
		auth = c.GetHeader("Authorization")

		logger.Log.Info("COOK HEADER AUTH:", zap.String("auth", auth))
		if auth == "" {
			return nil, http.ErrNoCookie
		}
	}

	p, err := service.Repository.ParsePayload(auth)

	logger.Log.Info("RESULT AUTH:", zap.String("auth", auth), zap.String("userid", p.UserID))

	return service.Repository.ParsePayload(auth)
}

func (service *AuthService) SaveAuthorization(UserID string, c *gin.Context) (string, error) {
	jwt, err := service.Repository.BuildJWTString(UserID)

	if err != nil {
		return "", err
	}

	c.SetCookie("Authorization", jwt, 86400, "/", "", false, true)

	return jwt, nil
}

func (service *AuthService) GetOrCreateAndSaveAuthorization(c *gin.Context) (*repository.Claims, error) {
	auth, err := service.GetAuthorization(c)

	if err == nil {
		return auth, err
	}

	return service.CreateAndSaveAuthorization(c)
}

func (service *AuthService) CreateAndSaveAuthorization(c *gin.Context) (*repository.Claims, error) {
	UserID, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	jwt, err := service.SaveAuthorization(UserID.String(), c)
	if err != nil {
		return nil, err
	}

	return service.Repository.ParsePayload(jwt)
}

func NewAuthService(config *config.AppConfig) *AuthService {
	return &AuthService{Repository: repository.NewAuthRepository(config)}
}
