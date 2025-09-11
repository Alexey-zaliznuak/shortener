package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type DBFlagsInitialConfig struct {
	DatabaseDSN *string
}

type FlagsInitialConfig struct {
	StoragePath    *string
	StartupAddress *string
	BaseURL        *string

	DB *DBFlagsInitialConfig
}

type DBConfig struct {
	DatabaseDSN string
	StoragePath string
}

type AuthConfig struct {
	TokenLifeTimeHours int
	TokenSecretKey     string
}

type AppConfig struct {
	LoggingLevel string

	DB   DBConfig
	Auth AuthConfig

	Server struct {
		BaseURL          string
		Address          string
		ShortLinksLength int
	}
}

type AppConfigBuilder struct {
	config      *AppConfig
	flagsConfig *FlagsInitialConfig
	Errors      []error
}

var (
	defaultStoragePath        = "storage.json"
	defaultShortLinksLength   = 8
	defaultStartupAddress     = "localhost:8080"
	defaultLoggingLevel       = "info"
	defaultTokenLifeTimeHours = 24
	defaultTokenSecretKey     = "superTokenSecretKey"
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

	b.config.Server.Address = b.loadStringVariableFromEnv("SERVER_ADDRESS", &def)

	return b
}

func (b *AppConfigBuilder) WithDatabaseDSN() *AppConfigBuilder {
	def := ""

	if b.flagsConfig.DB != nil && b.flagsConfig.DB.DatabaseDSN != nil && *b.flagsConfig.DB.DatabaseDSN != "" {
		def = *b.flagsConfig.DB.DatabaseDSN
	}

	b.config.DB.DatabaseDSN = b.loadStringVariableFromEnv("DATABASE_CONN_STRING", &def)
	return b
}

func (b *AppConfigBuilder) WithStoragePath() *AppConfigBuilder {
	def := defaultStoragePath

	if b.flagsConfig.StoragePath != nil && *b.flagsConfig.StoragePath != "" {
		def = *b.flagsConfig.StoragePath
	}

	b.config.DB.StoragePath = b.loadStringVariableFromEnv("FILE_STORAGE_PATH", &def)

	return b
}

func (b *AppConfigBuilder) WithTokenLifeTime() *AppConfigBuilder {
	def := defaultTokenLifeTimeHours

	b.config.Auth.TokenLifeTimeHours = b.loadIntVariableFromEnv("AUTH_TOKEN_LIFE_TIME_HOURS", &def)

	return b
}

func (b *AppConfigBuilder) WithTokenSecretKey() *AppConfigBuilder {
	b.config.Auth.TokenSecretKey = b.loadStringVariableFromEnv("AUTH_TOKEN_SECRET_KEY", &defaultTokenSecretKey)

	return b
}

func (b *AppConfigBuilder) WithBaseURL() *AppConfigBuilder {
	b.config.Server.BaseURL = b.loadStringVariableFromEnv("BASE_URL", b.flagsConfig.BaseURL)
	return b
}

func (b *AppConfigBuilder) WithShortLinksLength() *AppConfigBuilder {
	b.config.Server.ShortLinksLength = b.loadIntVariableFromEnv("SHORT_LINKS_LENGTH", &defaultShortLinksLength)
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
		BaseURL:        flag.String("b", "", "short links url prefix"),
		DB: &DBFlagsInitialConfig{
			DatabaseDSN: flag.String("d", "", "Database DSN"),
		},
		StoragePath: flag.String("f", "", "storage path to save dump and load all data"),
	}
}

var GetConfig = func(flagsConfig *FlagsInitialConfig) (*AppConfig, error) {
	return NewAppConfigBuilder(flagsConfig).
		WithBaseURL().
		WithDatabaseDSN().
		WithStoragePath().
		WithStartupAddress().
		WithShortLinksLength().
		WithLoggingLevel().
		WithTokenSecretKey().
		WithTokenLifeTime().
		Build()
}
