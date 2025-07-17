package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/Alexey-zaliznuak/shortener/internal/config/db"
)

type AppConfig struct {
	db.DatabaseConfig
	Port           int
	StartupAddress string

	ShortLinksLength int

	ShortLinksURLPrefix string
}

var (
	StartupAddressFlag      = flag.String("a", "", "startup address")
	ShortLinksURLPrefixFlag = flag.String("b", "", "short links url prefix")
)

var Config = func() AppConfig {
	flag.Parse()

	ServerStartupAddress := *StartupAddressFlag
	ShortLinksURLPrefix := *ShortLinksURLPrefixFlag

	if ServerStartupAddress == "" {
		ServerStartupAddress = os.Getenv("SERVER_STARTUP_ADDRESS")
	}

	if ServerStartupAddress == "" {
		ServerStartupAddress = "localhost:8080"
		fmt.Printf("configuration warning: 'SERVER_STARTUP_ADDRESS' not specified, use based on port: %s\n", ServerStartupAddress)
	}

	if ShortLinksURLPrefix == "" {
		ShortLinksURLPrefix = os.Getenv("SHORT_LINKS_URL_PREFIX")
	}

	if ShortLinksURLPrefix == "" {
		fmt.Println("configuration warning: 'SHORT_LINKS_URL_PREFIX' not specified")
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
		StartupAddress: ServerStartupAddress,

		ShortLinksLength:    ShortLinksLength,
		ShortLinksURLPrefix: ShortLinksURLPrefix,

		DatabaseConfig: db.LoadedDatabaseConfig,
	}
}()
