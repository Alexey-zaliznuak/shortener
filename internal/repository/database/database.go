package database

import (
	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Client = func() *gorm.DB {
	var err error
	var db *gorm.DB

	if config.Config.DatabaseConfig.InMemory {
		db, err = gorm.Open(sqlite.Open(config.Config.DatabaseConfig.Uri), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	} else {
		db, err = gorm.Open(postgres.Open(config.Config.DatabaseConfig.Uri), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	}
	if err != nil {
		panic("database error: cannot connect to the database: " + err.Error())
	}
	return db
}()
