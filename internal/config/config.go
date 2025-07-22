package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type FlagsInitialConfig struct {
	StartupAddress      *string
	ShortLinksURLPrefix *string
}

type AppConfig struct {
	Port                int
	StartupAddress      string
	ShortLinksLength    int
	ShortLinksURLPrefix string
}

type AppConfigBuilder struct {
	config *AppConfig
}

const (
    defaultShortLinksLength = 8
    defaultStartupAddress   = "localhost:8080"
)

func NewAppConfigBuilder() *AppConfigBuilder {
	return &AppConfigBuilder{
		config: &AppConfig{},
	}
}

func (b *AppConfigBuilder) WithStartupAddress(flagsConfig *FlagsInitialConfig) *AppConfigBuilder {
	startupAddress := ""
	if flagsConfig.StartupAddress != nil {
		startupAddress = *flagsConfig.StartupAddress
	}
	if startupAddress == "" {
		startupAddress = os.Getenv("SERVER_STARTUP_ADDRESS")
	}
	if startupAddress == "" {
		startupAddress = defaultStartupAddress
		Logger.Printf("configuration warning: 'SERVER_STARTUP_ADDRESS' not specified, using default: %s\n", startupAddress)
	}
	b.config.StartupAddress = startupAddress
	return b
}

func (b *AppConfigBuilder) WithShortLinksURLPrefix(flagsConfig *FlagsInitialConfig) *AppConfigBuilder {
	shortLinksURLPrefix := ""
	if flagsConfig.ShortLinksURLPrefix != nil {
		shortLinksURLPrefix = *flagsConfig.ShortLinksURLPrefix
	}
	if shortLinksURLPrefix == "" {
		shortLinksURLPrefix = os.Getenv("SHORT_LINKS_URL_PREFIX")
	}
	if shortLinksURLPrefix == "" {
		Logger.Println("configuration warning: 'SHORT_LINKS_URL_PREFIX' not specified")
	}
	b.config.ShortLinksURLPrefix = shortLinksURLPrefix
	return b
}

func (b *AppConfigBuilder) WithShortLinksLength() *AppConfigBuilder {
	shortLinksLength, err := strconv.Atoi(os.Getenv("SHORT_LINKS_LENGTH"))
	if err != nil {
		shortLinksLength = defaultShortLinksLength
		Logger.Printf("configuration warning: 'SHORT_LINKS_LENGTH' not specified, using default: %d\n", shortLinksLength)
	}
	b.config.ShortLinksLength = shortLinksLength
	return b
}

func (b *AppConfigBuilder) Build() *AppConfig {
	return b.config
}

var Logger = log.Default()

func CreateFLagsInitialConfig() *FlagsInitialConfig {
	return &FlagsInitialConfig{
		StartupAddress:      flag.String("a", "", "startup address"),
		ShortLinksURLPrefix: flag.String("b", "", "short links url prefix"),
	}
}

var GetConfig = func(flagsConfig *FlagsInitialConfig) *AppConfig {
	return NewAppConfigBuilder().
		WithStartupAddress(flagsConfig).
		WithShortLinksURLPrefix(flagsConfig).
		WithShortLinksLength().
		Build()
}
