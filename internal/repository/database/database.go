package database

import (
	"fmt"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Client *gorm.DB = nil

func init() {
	var err error

	if config.GetConfig().DatabaseConfig.InMemory {
		Client, err = gorm.Open(sqlite.Open(config.GetConfig().DatabaseConfig.URI), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	} else {
		Client, err = gorm.Open(postgres.Open(config.GetConfig().DatabaseConfig.URI), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	}
	if err != nil {
		panic(fmt.Errorf("database error: cannot connect to the database: %v", err.Error()))
	}
}
