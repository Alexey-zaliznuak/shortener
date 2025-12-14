// Package config предоставляет функциональность для управления конфигурацией приложения.
// Конфигурация загружается из флагов командной строки, переменных окружения и JSON-файла,
// с приоритетом: переменные окружения > флаги > JSON-файл.
//
//go:generate go run ../../cmd/reset
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
)

// JSONConfig содержит конфигурацию из JSON-файла.
type JSONConfig struct {
	// ServerAddress содержит адрес сервера (аналог флага -a и SERVER_ADDRESS).
	ServerAddress string `json:"server_address"`
	// BaseURL содержит базовый URL для коротких ссылок (аналог флага -b и BASE_URL).
	BaseURL string `json:"base_url"`
	// FileStoragePath содержит путь к файлу хранилища (аналог флага -f и FILE_STORAGE_PATH).
	FileStoragePath string `json:"file_storage_path"`
	// DatabaseDSN содержит строку подключения к БД (аналог флага -d и DATABASE_DSN).
	DatabaseDSN string `json:"database_dsn"`
	// EnableHTTPS включает HTTPS режим (аналог флага -s и ENABLE_HTTPS).
	EnableHTTPS bool `json:"enable_https"`
	// LoggingLevel содержит уровень логирования (аналог LOGGING_LEVEL).
	LoggingLevel string `json:"logging_level"`
	// ShortLinksLength содержит длину коротких ссылок (аналог SHORT_LINKS_LENGTH).
	ShortLinksLength int `json:"short_links_length"`
	// TokenLifeTimeHours содержит время жизни токена в часах (аналог AUTH_TOKEN_LIFE_TIME_HOURS).
	TokenLifeTimeHours int `json:"token_life_time_hours"`
	// TokenSecretKey содержит секретный ключ для токенов (аналог AUTH_TOKEN_SECRET_KEY).
	TokenSecretKey string `json:"token_secret_key"`
	// AuditURL содержит URL для аудита (аналог флага -audit-url).
	AuditURL string `json:"audit_url"`
	// AuditFile содержит путь к файлу аудита (аналог флага -audit-file).
	AuditFile string `json:"audit_file"`
}

// LoadJSONConfig загружает конфигурацию из JSON-файла.
// Возвращает nil, если путь пуст или файл не найден.
func LoadJSONConfig(configPath string) (*JSONConfig, error) {
	if configPath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg JSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// getConfigPath возвращает путь к файлу конфигурации из флага или переменной окружения.
func getConfigPath(flagConfigPath *string) string {
	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath
	}
	if flagConfigPath != nil && *flagConfigPath != "" {
		return *flagConfigPath
	}
	return ""
}

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

	// EnableHTTPS включает HTTPS режим для сервера.
	EnableHTTPS *bool

	// ConfigPath содержит путь к файлу конфигурации JSON.
	ConfigPath *string
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
// generate:reset
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
		// EnableHTTPS включает HTTPS режим для сервера.
		EnableHTTPS bool
	}
}

// AppConfigBuilder предоставляет паттерн Builder для построения конфигурации приложения.
type AppConfigBuilder struct {
	config      *AppConfig
	flagsConfig *FlagsInitialConfig
	jsonConfig  *JSONConfig

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
func NewAppConfigBuilder(flagsConfig *FlagsInitialConfig, jsonConfig *JSONConfig) *AppConfigBuilder {
	return &AppConfigBuilder{
		config:      &AppConfig{},
		flagsConfig: flagsConfig,
		jsonConfig:  jsonConfig,
	}
}

// WithStartupAddress устанавливает адрес запуска сервера из переменной окружения SERVER_ADDRESS,
func (b *AppConfigBuilder) WithStartupAddress() *AppConfigBuilder {
	def := defaultStartupAddress

	if b.jsonConfig != nil && b.jsonConfig.ServerAddress != "" {
		def = b.jsonConfig.ServerAddress
	}

	if b.flagsConfig.StartupAddress != nil && *b.flagsConfig.StartupAddress != "" {
		def = *b.flagsConfig.StartupAddress
	}

	b.config.Server.Address = b.loadStringVariableFromEnv("SERVER_ADDRESS", &def)

	return b
}

// WithDatabaseDSN устанавливает строку подключения к базе данных из переменной окружения
func (b *AppConfigBuilder) WithDatabaseDSN() *AppConfigBuilder {
	def := ""

	if b.jsonConfig != nil && b.jsonConfig.DatabaseDSN != "" {
		def = b.jsonConfig.DatabaseDSN
	}

	if b.flagsConfig.DB != nil && b.flagsConfig.DB.DatabaseDSN != nil && *b.flagsConfig.DB.DatabaseDSN != "" {
		def = *b.flagsConfig.DB.DatabaseDSN
	}

	b.config.DB.DatabaseDSN = b.loadStringVariableFromEnv("DATABASE_CONN_STRING", &def)
	return b
}

// WithStoragePath устанавливает путь к файлу хранилища из переменной окружения
func (b *AppConfigBuilder) WithStoragePath() *AppConfigBuilder {
	def := defaultStoragePath

	if b.jsonConfig != nil && b.jsonConfig.FileStoragePath != "" {
		def = b.jsonConfig.FileStoragePath
	}

	if b.flagsConfig.StoragePath != nil && *b.flagsConfig.StoragePath != "" {
		def = *b.flagsConfig.StoragePath
	}

	b.config.DB.StoragePath = b.loadStringVariableFromEnv("FILE_STORAGE_PATH", &def)

	return b
}

// WithTokenLifeTime устанавливает время жизни токена из переменной окружения
func (b *AppConfigBuilder) WithTokenLifeTime() *AppConfigBuilder {
	def := defaultTokenLifeTimeHours

	if b.jsonConfig != nil && b.jsonConfig.TokenLifeTimeHours > 0 {
		def = b.jsonConfig.TokenLifeTimeHours
	}

	b.config.Auth.TokenLifeTimeHours = b.loadIntVariableFromEnv("AUTH_TOKEN_LIFE_TIME_HOURS", &def)

	return b
}

// WithTokenSecretKey устанавливает секретный ключ токена из переменной окружения
func (b *AppConfigBuilder) WithTokenSecretKey() *AppConfigBuilder {
	def := defaultTokenSecretKey

	if b.jsonConfig != nil && b.jsonConfig.TokenSecretKey != "" {
		def = b.jsonConfig.TokenSecretKey
	}

	b.config.Auth.TokenSecretKey = b.loadStringVariableFromEnv("AUTH_TOKEN_SECRET_KEY", &def)

	return b
}

// WithBaseURL устанавливает базовый URL из переменной окружения BASE_URL
func (b *AppConfigBuilder) WithBaseURL() *AppConfigBuilder {
	var def *string

	if b.jsonConfig != nil && b.jsonConfig.BaseURL != "" {
		def = &b.jsonConfig.BaseURL
	}

	if b.flagsConfig.BaseURL != nil && *b.flagsConfig.BaseURL != "" {
		def = b.flagsConfig.BaseURL
	}

	b.config.Server.BaseURL = b.loadStringVariableFromEnv("BASE_URL", def)
	return b
}

// WithShortLinksLength устанавливает длину коротких ссылок из переменной окружения
func (b *AppConfigBuilder) WithShortLinksLength() *AppConfigBuilder {
	def := defaultShortLinksLength

	if b.jsonConfig != nil && b.jsonConfig.ShortLinksLength > 0 {
		def = b.jsonConfig.ShortLinksLength
	}

	b.config.Server.ShortLinksLength = b.loadIntVariableFromEnv("SHORT_LINKS_LENGTH", &def)
	return b
}

// WithLoggingLevel устанавливает уровень логирования из переменной окружения
func (b *AppConfigBuilder) WithLoggingLevel() *AppConfigBuilder {
	def := defaultLoggingLevel

	if b.jsonConfig != nil && b.jsonConfig.LoggingLevel != "" {
		def = b.jsonConfig.LoggingLevel
	}

	b.config.LoggingLevel = b.loadStringVariableFromEnv("LOGGING_LEVEL", &def)
	return b
}

// WithAuditFile устанавливает путь к файлу аудита из переменной окружения
func (b *AppConfigBuilder) WithAuditFile() *AppConfigBuilder {

	if b.jsonConfig != nil && b.jsonConfig.AuditFile != "" {
		b.config.Audit.AuditFile = b.jsonConfig.AuditFile
	}

	if b.flagsConfig.AuditFile != nil && *b.flagsConfig.AuditFile != "" {
		b.config.Audit.AuditFile = *b.flagsConfig.AuditFile
	}

	return b
}

// WithAuditURL устанавливает URL аудита из переменной окружения
func (b *AppConfigBuilder) WithAuditURL() *AppConfigBuilder {

	if b.jsonConfig != nil && b.jsonConfig.AuditURL != "" {
		b.config.Audit.AuditURL = b.jsonConfig.AuditURL
	}

	if b.flagsConfig.AuditURL != nil && *b.flagsConfig.AuditURL != "" {
		b.config.Audit.AuditURL = *b.flagsConfig.AuditURL
	}

	return b
}

// WithEnableHTTPS устанавливает режим HTTPS из переменной окружения ENABLE_HTTPS,
func (b *AppConfigBuilder) WithEnableHTTPS() *AppConfigBuilder {
	def := false

	if b.jsonConfig != nil && b.jsonConfig.EnableHTTPS {
		def = true
	}

	if b.flagsConfig.EnableHTTPS != nil && *b.flagsConfig.EnableHTTPS {
		def = true
	}

	b.config.Server.EnableHTTPS = b.loadBoolVariableFromEnv("ENABLE_HTTPS", &def)

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

// loadBoolVariableFromEnv загружает булево значение из переменной окружения.
// Значение преобразуется из строки в bool. Поддерживаемые значения: "true", "1", "false", "0", "".
func (b *AppConfigBuilder) loadBoolVariableFromEnv(envName string, Default *bool) bool {
	envValue := os.Getenv(envName)

	if envValue == "" {
		if Default != nil {
			return *Default
		}
		return false
	}

	boolValue, err := strconv.ParseBool(envValue)
	if err != nil {
		b.Errors = append(b.Errors, fmt.Errorf("configuration error: could not convert %s to bool: %w", envName, err))
		return false
	}

	return boolValue
}

// CreateFLagsInitialConfig создает и инициализирует FlagsInitialConfig с флагами командной строки.
// Флаги должны быть распарсены с помощью flag.Parse() перед использованием.
func CreateFLagsInitialConfig() *FlagsInitialConfig {
	configPath := flag.String("c", "", "path to JSON config file")
	flag.StringVar(configPath, "config", "", "path to JSON config file")

	return &FlagsInitialConfig{
		StartupAddress: flag.String("a", "", "startup address"),
		BaseURL:        flag.String("b", "", "short links url prefix"),
		DB: &DBFlagsInitialConfig{
			DatabaseDSN: flag.String("d", "", "Database DSN"),
		},
		StoragePath: flag.String("f", "", "storage path to save dump and load all data"),
		AuditURL:    flag.String("audit-url", "", "audit HTTP endpoint URL"),
		AuditFile:   flag.String("audit-file", "", "audit log file path"),
		EnableHTTPS: flag.Bool("s", false, "enable HTTPS mode"),
		ConfigPath:  configPath,
	}
}

// GetConfig создает полную конфигурацию приложения из флагов, переменных окружения и JSON-файла.
// Возвращает готовую конфигурацию или ошибки, возникшие при ее построении.
var GetConfig = func(flagsConfig *FlagsInitialConfig) (*AppConfig, error) {
	// Загружаем JSON конфигурацию
	configPath := getConfigPath(flagsConfig.ConfigPath)
	jsonConfig, err := LoadJSONConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON config: %w", err)
	}

	return NewAppConfigBuilder(flagsConfig, jsonConfig).
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
		WithEnableHTTPS().
		Build()
}
