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
}

var Config = func() AppConfig {
	ShortLinksLength, err := strconv.Atoi(os.Getenv("SHORT_LINKS_LENGTH"))

	if err != nil {
		ShortLinksLength = 8
		fmt.Println("configuration warning: 'SHORT_LINKS_LENGTH' not specified")
	}

	if ShortLinksLength < 0 {
		panic(fmt.Sprintf("configuration error: 'SHORT_LINKS_LENGTH' has invalid value %d", ShortLinksLength))
	}

	return AppConfig{
		DatabaseConfig:   db.LoadedDatabaseConfig,
		ShortLinksLength: ShortLinksLength,
	}
}()
