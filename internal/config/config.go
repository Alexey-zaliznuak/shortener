package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Alexey-zaliznuak/shortener/internal/config/db"
)

type AppConfig struct {
	db.DatabaseConfig
	Port                 int
	ServerStartupAddress string

	ShortLinksLength int

	ShortLinksURLPrefix string
}

var Config = func() AppConfig {
	Port, err := strconv.Atoi(os.Getenv("PORT"))

	if err != nil {
		Port = 8080
		fmt.Printf("configuration warning: 'PORT' not specified, use default: %d\n", Port)
	}

	ServerStartupAddress := os.Getenv("SERVER_STARTUP_ADDRESS")

	if ServerStartupAddress == "" {
		ServerStartupAddress = fmt.Sprintf("localhost:%d", Port)
		fmt.Printf("configuration warning: 'SERVER_STARTUP_ADDRESS' not specified, use based on port: %s\n", ServerStartupAddress)
	}

	ShortLinksURLPrefix := os.Getenv("SHORT_LINKS_URL_PREFIX")

	if ShortLinksURLPrefix == "" {
		fmt.Printf("configuration warning: 'SHORT_LINKS_URL_PREFIX' not specified")
	}

	ShortLinksLength, err := strconv.Atoi(os.Getenv("SHORT_LINKS_LENGTH"))

	if err != nil {
		ShortLinksLength = 8
		fmt.Printf("configuration warning: 'SHORT_LINKS_LENGTH' not specified, use default: %d\n", ShortLinksLength)
	}

	if ShortLinksLength < 0 {
		panic(fmt.Sprintf("configuration error: 'SHORT_LINKS_LENGTH' has invalid value %d", ShortLinksLength))
	}

	return AppConfig{
		Port:                 Port,
		ServerStartupAddress: ServerStartupAddress,

		ShortLinksLength:    ShortLinksLength,
		ShortLinksURLPrefix: ShortLinksURLPrefix,

		DatabaseConfig: db.LoadedDatabaseConfig,
	}
}()
