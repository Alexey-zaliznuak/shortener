package repository

import (
	"errors"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/golang-jwt/jwt/v4"
)

var ErrTokenValidation = errors.New("invalid token signature")

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type AuthRepository struct {
	config *config.AppConfig
}

func (repository *AuthRepository) BuildJWTString(UserID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(
				time.Hour * time.Duration(repository.config.Auth.TokenLifeTimeHours),
			)),
		},
		// собственное утверждение
		UserID: UserID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(repository.config.Auth.TokenSecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (repository *AuthRepository) ParsePayload(token string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenValidation
		}
		return []byte(repository.config.Auth.TokenSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}

func NewAuthRepository(config *config.AppConfig) *AuthRepository {
	return &AuthRepository{config: config}
}
