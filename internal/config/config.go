package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/Alexey-zaliznuak/shortener/internal/config/db"
)

type AppConfig struct {
	db.DatabaseConfig
	Port                int
	StartupAddress      string
	ShortLinksLength    int
	ShortLinksURLPrefix string
}

var (
	startupAddressFlag      = flag.String("a", "", "startup address")
	shortLinksURLPrefixFlag = flag.String("b", "", "short links url prefix")

	config   *AppConfig
	initOnce sync.Once
	initErr  error
)

func InitConfig() error {
	initOnce.Do(func() {
		flag.Parse()

		serverStartupAddress := *startupAddressFlag
		shortLinksURLPrefix := *shortLinksURLPrefixFlag

		if serverStartupAddress == "" {
			serverStartupAddress = os.Getenv("SERVER_STARTUP_ADDRESS")
		}
		if serverStartupAddress == "" {
			serverStartupAddress = "localhost:8080"
			fmt.Printf("configuration warning: 'SERVER_STARTUP_ADDRESS' not specified, using default: %s\n", serverStartupAddress)
		}

		if shortLinksURLPrefix == "" {
			shortLinksURLPrefix = os.Getenv("SHORT_LINKS_URL_PREFIX")
		}
		if shortLinksURLPrefix == "" {
			fmt.Println("configuration warning: 'SHORT_LINKS_URL_PREFIX' not specified")
		}

		shortLinksLength, err := strconv.Atoi(os.Getenv("SHORT_LINKS_LENGTH"))
		if err != nil {
			shortLinksLength = 8
			fmt.Printf("configuration warning: 'SHORT_LINKS_LENGTH' not specified, using default: %d\n", shortLinksLength)
		}
		if shortLinksLength < 0 {
			initErr = fmt.Errorf("configuration error: 'SHORT_LINKS_LENGTH' has invalid value %d", shortLinksLength)
			return
		}

		config = &AppConfig{
			StartupAddress:      serverStartupAddress,
			ShortLinksLength:    shortLinksLength,
			ShortLinksURLPrefix: shortLinksURLPrefix,
			DatabaseConfig:      db.CreateDatabaseConfig(),
		}
	})
	return initErr
}

func GetConfig() *AppConfig {
	if err := InitConfig(); err != nil {
		panic(fmt.Sprintf("failed to initialize config: %v", err))
	}
	return config
}
