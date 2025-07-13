package config

import "github.com/Alexey-zaliznuak/shortener/internal/config/db"

type AppConfig struct {
	db.DatabaseConfig
}

var Config = func() AppConfig {
	return AppConfig{DatabaseConfig: db.LoadedDatabaseConfig}
}()
