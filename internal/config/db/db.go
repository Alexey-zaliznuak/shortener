package db

import (
	"fmt"
	"os"
)

type DatabaseConfig struct {
	URI      string
	InMemory bool
}

func CreateDatabaseConfig() DatabaseConfig {
	var URI = os.Getenv("DATABASE_URI")
	if URI == "" {
		fmt.Println("configuration warning: db: 'DATABASE_URI' not provided in env, use in memory sqlite db")
		return DatabaseConfig{URI: "file::memory:?cache=shared", InMemory: true}
	}
	return DatabaseConfig{URI: URI, InMemory: false}
}
