package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type FlagsInitialConfig struct {
	StartupAddress *string
	BaseUrl        *string
}

type AppConfig struct {
	LoggingLevel     string `json:"loggingLevel"`
	ServerAddress    string `json:"serverAddress"`
	BaseURL          string `json:"baseURL"`
	ShortLinksLength int    `json:"shortLinksLength"`
}

type AppConfigBuilder struct {
	config      *AppConfig
	flagsConfig *FlagsInitialConfig
	Errors      []error
}

var (
	defaultShortLinksLength = 8
	defaultStartupAddress   = "localhost:8080"
	defaultBaseUrl          = "http://localhost:8080/"
	defaultLoggingLevel     = "info"
)

func NewAppConfigBuilder(flagsConfig *FlagsInitialConfig) *AppConfigBuilder {
	return &AppConfigBuilder{
		config: &AppConfig{}, flagsConfig: flagsConfig,
	}
}

func (b *AppConfigBuilder) WithStartupAddress() *AppConfigBuilder {
	def := defaultStartupAddress

	if b.flagsConfig.StartupAddress != nil && *b.flagsConfig.StartupAddress != "" {
		def = *b.flagsConfig.StartupAddress
	}

	b.config.ServerAddress = b.loadStringVariableFromEnv("SERVER_ADDRESS", &def)

	return b
}

func (b *AppConfigBuilder) WithBaseUrl() *AppConfigBuilder {
	def := defaultBaseUrl

	if b.flagsConfig.BaseUrl != nil && *b.flagsConfig.BaseUrl != "" {
		def = *b.flagsConfig.BaseUrl
	}

	b.config.BaseURL = b.loadStringVariableFromEnv("BASE_URL", &def)

	return b
}

func (b *AppConfigBuilder) WithShortLinksLength() *AppConfigBuilder {
	b.config.ShortLinksLength = b.loadIntVariableFromEnv("SHORT_LINKS_LENGTH", &defaultShortLinksLength)
	return b
}

func (b *AppConfigBuilder) WithLoggingLevel() *AppConfigBuilder {
	b.config.LoggingLevel = b.loadStringVariableFromEnv("LOGGING_LEVEL", &defaultLoggingLevel)
	return b
}

func (b *AppConfigBuilder) Build() (*AppConfig, error) {
	return b.config, errors.Join(b.Errors...)
}

func (b *AppConfigBuilder) loadStringVariableFromEnv(envName string, Default *string) string {
	value := os.Getenv(envName)

	if value == "" && Default != nil {
		value = *Default
	}

	if value == "" {
		b.Errors = append(b.Errors, fmt.Errorf("configuration error: '%s' not specified", envName))
	}

	return value
}

func (b *AppConfigBuilder) loadIntVariableFromEnv(envName string, Default *int) int {
	stringedDefault := strconv.Itoa(*Default)
	value := b.loadStringVariableFromEnv(envName, &stringedDefault)

	if value == "" {
		return 0
	}

	numericValue, err := strconv.Atoi(value)

	if err != nil {
		b.Errors = append(b.Errors, fmt.Errorf("configuration error: could not convert %s to int: %w", envName, err))
	}

	return numericValue
}

func CreateFLagsInitialConfig() *FlagsInitialConfig {
	return &FlagsInitialConfig{
		StartupAddress: flag.String("a", "", "startup address"),
		BaseUrl:        flag.String("b", "", "short links url prefix"),
	}
}

var GetConfig = func(flagsConfig *FlagsInitialConfig) (*AppConfig, error) {
	return NewAppConfigBuilder(flagsConfig).
		WithStartupAddress().
		WithBaseUrl().
		WithShortLinksLength().
		WithLoggingLevel().
		Build()
}
