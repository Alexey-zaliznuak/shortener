package database

import (
	"fmt"
	"sync"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	client    *gorm.DB
	initOnce  sync.Once
)

// GetClient возвращает инициализированное подключение к БД
func GetClient() *gorm.DB {
	initOnce.Do(func() {
		var err error
		dbConf := config.GetConfig().DatabaseConfig

		if dbConf.InMemory {
			client, err = gorm.Open(sqlite.Open(dbConf.URI), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})
		} else {
			client, err = gorm.Open(postgres.Open(dbConf.URI), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})
		}

		client.AutoMigrate(&model.Link{})

		if err != nil {
			panic(fmt.Errorf("database error: cannot connect to the database: %v", err))
		}
	})
	return client
}
