package db

import (
	"fmt"
	"os"
)

type DatabaseConfig struct {
	Uri      string
	InMemory bool
}

var LoadedDatabaseConfig = func() DatabaseConfig {
	var Uri = os.Getenv("DATABASE_URI")
	if Uri == "" {
		fmt.Println("configuration warning: db: 'DATABASE_URI' not provided in env, use in memory sqlite db")
		return DatabaseConfig{Uri: "file::memory:?cache=shared", InMemory: true}
		// panic("configuration error: db: 'DATABASE_URI' not provided in env")
	}
	return DatabaseConfig{Uri: Uri, InMemory: false}
}()
