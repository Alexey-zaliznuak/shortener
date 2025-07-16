package db

import (
	"fmt"
	"os"
)

type DatabaseConfig struct {
	URI      string
	InMemory bool
}

var LoadedDatabaseConfig = func() DatabaseConfig {
	var URI = os.Getenv("DATABASE_URI")
	if URI == "" {
		fmt.Println("configuration warning: db: 'DATABASE_URI' not provided in env, use in memory sqlite db")
		return DatabaseConfig{URI: "file::memory:?cache=shared", InMemory: true}
		// panic("configuration error: db: 'DATABASE_URI' not provided in env")
	}
	return DatabaseConfig{URI: URI, InMemory: false}
}()
