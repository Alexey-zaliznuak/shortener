// Package config предоставляет функциональность для управления конфигурацией приложения.
// Конфигурация загружается из флагов командной строки и переменных окружения,
// с приоритетом переменных окружения над флагами.
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
)

// DBFlagsInitialConfig содержит начальную конфигурацию базы данных из флагов.
type DBFlagsInitialConfig struct {
	// DatabaseDSN содержит строку подключения к базе данных.
	DatabaseDSN *string
}

// FlagsInitialConfig содержит начальную конфигурацию приложения из флагов командной строки.
type FlagsInitialConfig struct {
	// StoragePath содержит путь к файлу хранилища данных.
	StoragePath *string
	// StartupAddress содержит адрес запуска сервера.
	StartupAddress *string
	// BaseURL содержит базовый URL для коротких ссылок.
	BaseURL *string

	// DB содержит конфигурацию базы данных.
	DB *DBFlagsInitialConfig

	// AuditURL содержит URL HTTP-эндпоинта для аудита.
	AuditURL *string
	// AuditFile содержит путь к файлу логов аудита.
	AuditFile *string
}

// DBConfig содержит конфигурацию базы данных и хранилища.
type DBConfig struct {
	// DatabaseDSN содержит строку подключения к базе данных.
	DatabaseDSN string
	// StoragePath содержит путь к файлу хранилища данных.
	StoragePath string
}

// AuthConfig содержит конфигурацию аутентификации.
type AuthConfig struct {
	// TokenLifeTimeHours содержит время жизни токена в часах.
	TokenLifeTimeHours int
	// TokenSecretKey содержит секретный ключ для подписи токенов.
	TokenSecretKey string
}

// AppConfig содержит полную конфигурацию приложения.
type AppConfig struct {
	// LoggingLevel содержит уровень логирования.
	LoggingLevel string

	// DB содержит конфигурацию базы данных.
	DB DBConfig
	// Auth содержит конфигурацию аутентификации.
	Auth AuthConfig

	// Audit содержит конфигурацию аудита.
	Audit struct {
		// AuditURL содержит URL для отправки событий аудита.
		AuditURL string
		// AuditFile содержит путь к файлу логов аудита.
		AuditFile string
	}

	// Server содержит конфигурацию сервера.
	Server struct {
		// BaseURL содержит базовый URL для генерации коротких ссылок.
		BaseURL string
		// Address содержит адрес, на котором запускается сервер.
		Address string
		// ShortLinksLength содержит длину генерируемых коротких ссылок.
		ShortLinksLength int
	}
}

// AppConfigBuilder предоставляет паттерн Builder для построения конфигурации приложения.
type AppConfigBuilder struct {
	config      *AppConfig
	flagsConfig *FlagsInitialConfig
	// Errors содержит список ошибок, возникших при построении конфигурации.
	Errors []error
}

var (
	defaultStoragePath        = "storage.json"
	defaultShortLinksLength   = 8
	defaultStartupAddress     = "localhost:8080"
	defaultLoggingLevel       = "info"
	defaultTokenLifeTimeHours = 24
	defaultTokenSecretKey     = "superTokenSecretKey"
)

// NewAppConfigBuilder создает новый экземпляр AppConfigBuilder с указанной начальной конфигурацией флагов.
func NewAppConfigBuilder(flagsConfig *FlagsInitialConfig) *AppConfigBuilder {
	return &AppConfigBuilder{
		config: &AppConfig{}, flagsConfig: flagsConfig,
	}
}

// WithStartupAddress устанавливает адрес запуска сервера из переменной окружения SERVER_ADDRESS
// или флага командной строки. Если ни один не указан, используется значение по умолчанию.
func (b *AppConfigBuilder) WithStartupAddress() *AppConfigBuilder {
	def := defaultStartupAddress

	if b.flagsConfig.StartupAddress != nil && *b.flagsConfig.StartupAddress != "" {
		def = *b.flagsConfig.StartupAddress
	}

	b.config.Server.Address = b.loadStringVariableFromEnv("SERVER_ADDRESS", &def)

	return b
}

// WithDatabaseDSN устанавливает строку подключения к базе данных из переменной окружения
// DATABASE_CONN_STRING или флага командной строки.
func (b *AppConfigBuilder) WithDatabaseDSN() *AppConfigBuilder {
	def := ""

	if b.flagsConfig.DB != nil && b.flagsConfig.DB.DatabaseDSN != nil && *b.flagsConfig.DB.DatabaseDSN != "" {
		def = *b.flagsConfig.DB.DatabaseDSN
	}

	b.config.DB.DatabaseDSN = b.loadStringVariableFromEnv("DATABASE_CONN_STRING", &def)
	return b
}

// WithStoragePath устанавливает путь к файлу хранилища из переменной окружения
// FILE_STORAGE_PATH или флага командной строки. Если ни один не указан, используется значение по умолчанию.
func (b *AppConfigBuilder) WithStoragePath() *AppConfigBuilder {
	def := defaultStoragePath

	if b.flagsConfig.StoragePath != nil && *b.flagsConfig.StoragePath != "" {
		def = *b.flagsConfig.StoragePath
	}

	b.config.DB.StoragePath = b.loadStringVariableFromEnv("FILE_STORAGE_PATH", &def)

	return b
}

// WithTokenLifeTime устанавливает время жизни токена из переменной окружения
// AUTH_TOKEN_LIFE_TIME_HOURS. Если не указано, используется значение по умолчанию.
func (b *AppConfigBuilder) WithTokenLifeTime() *AppConfigBuilder {
	def := defaultTokenLifeTimeHours

	b.config.Auth.TokenLifeTimeHours = b.loadIntVariableFromEnv("AUTH_TOKEN_LIFE_TIME_HOURS", &def)

	return b
}

// WithTokenSecretKey устанавливает секретный ключ токена из переменной окружения
// AUTH_TOKEN_SECRET_KEY. Если не указано, используется значение по умолчанию.
func (b *AppConfigBuilder) WithTokenSecretKey() *AppConfigBuilder {
	b.config.Auth.TokenSecretKey = b.loadStringVariableFromEnv("AUTH_TOKEN_SECRET_KEY", &defaultTokenSecretKey)

	return b
}

// WithBaseURL устанавливает базовый URL из переменной окружения BASE_URL или флага командной строки.
func (b *AppConfigBuilder) WithBaseURL() *AppConfigBuilder {
	b.config.Server.BaseURL = b.loadStringVariableFromEnv("BASE_URL", b.flagsConfig.BaseURL)
	return b
}

// WithShortLinksLength устанавливает длину коротких ссылок из переменной окружения
// SHORT_LINKS_LENGTH. Если не указано, используется значение по умолчанию.
func (b *AppConfigBuilder) WithShortLinksLength() *AppConfigBuilder {
	b.config.Server.ShortLinksLength = b.loadIntVariableFromEnv("SHORT_LINKS_LENGTH", &defaultShortLinksLength)
	return b
}

// WithLoggingLevel устанавливает уровень логирования из переменной окружения
// LOGGING_LEVEL. Если не указано, используется значение по умолчанию.
func (b *AppConfigBuilder) WithLoggingLevel() *AppConfigBuilder {
	b.config.LoggingLevel = b.loadStringVariableFromEnv("LOGGING_LEVEL", &defaultLoggingLevel)
	return b
}

// WithAuditFile устанавливает путь к файлу аудита из флага командной строки.
func (b *AppConfigBuilder) WithAuditFile() *AppConfigBuilder {
	if b.flagsConfig.AuditFile != nil {
		b.config.Audit.AuditFile = *b.flagsConfig.AuditFile
	}

	return b
}

// WithAuditURL устанавливает URL аудита из флага командной строки.
func (b *AppConfigBuilder) WithAuditURL() *AppConfigBuilder {
	if b.flagsConfig.AuditURL != nil {
		b.config.Audit.AuditURL = *b.flagsConfig.AuditURL
	}

	return b
}

// Build завершает построение конфигурации и возвращает готовую AppConfig.
// Если при построении возникли ошибки, они возвращаются объединенными.
func (b *AppConfigBuilder) Build() (*AppConfig, error) {
	return b.config, errors.Join(b.Errors...)
}

// loadStringVariableFromEnv загружает строковое значение из переменной окружения.
// Если переменная окружения не установлена, используется значение по умолчанию.
// Если значение пустое и значение по умолчанию не указано, добавляется ошибка.
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

// loadIntVariableFromEnv загружает целочисленное значение из переменной окружения.
// Значение преобразуется из строки в int. Если преобразование не удается, добавляется ошибка.
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

// CreateFLagsInitialConfig создает и инициализирует FlagsInitialConfig с флагами командной строки.
// Флаги должны быть распарсены с помощью flag.Parse() перед использованием.
func CreateFLagsInitialConfig() *FlagsInitialConfig {
	return &FlagsInitialConfig{
		StartupAddress: flag.String("a", "", "startup address"),
		BaseURL:        flag.String("b", "", "short links url prefix"),
		DB: &DBFlagsInitialConfig{
			DatabaseDSN: flag.String("d", "", "Database DSN"),
		},
		StoragePath: flag.String("f", "", "storage path to save dump and load all data"),
		AuditURL:    flag.String("audit-url", "", "audit HTTP endpoint URL"),
		AuditFile:   flag.String("audit-file", "", "audit log file path"),
	}
}

// GetConfig создает полную конфигурацию приложения из флагов и переменных окружения.
// Возвращает готовую конфигурацию или ошибки, возникшие при ее построении.
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
		WithAuditFile().
		WithAuditURL().
		Build()
}