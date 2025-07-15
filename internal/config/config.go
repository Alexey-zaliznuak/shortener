package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Alexey-zaliznuak/shortener/internal/config/db"
)

type AppConfig struct {
	db.DatabaseConfig
	ShortLinksLength int
	Port             int
}

var Config = func() AppConfig {
	ShortLinksLength, err := strconv.Atoi(os.Getenv("SHORT_LINKS_LENGTH"))

	if err != nil {
		ShortLinksLength = 8
		fmt.Printf("configuration warning: 'SHORT_LINKS_LENGTH' not specified, use default: %d\n", ShortLinksLength)
	}

	Port, err := strconv.Atoi(os.Getenv("PORT"))

	if err != nil {
		Port = 8080
		fmt.Printf("configuration warning: 'PORT' not specified, use default: %d\n", Port)
	}

	if ShortLinksLength < 0 {
		panic(fmt.Sprintf("configuration error: 'SHORT_LINKS_LENGTH' has invalid value %d", ShortLinksLength))
	}

	return AppConfig{
		DatabaseConfig:   db.LoadedDatabaseConfig,
		ShortLinksLength: ShortLinksLength,
		Port:             Port,
	}
}()
