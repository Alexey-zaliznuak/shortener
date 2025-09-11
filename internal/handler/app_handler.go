package handler

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func checkHealth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.String(http.StatusInternalServerError, "Database connection not configured")
			return
		}

		err := db.PingContext(context.Background())

		if err != nil {
			c.String(http.StatusInternalServerError, "Database connection failed: %s", err.Error())
			return
		}

		c.String(http.StatusOK, "OK")
	}
}

func RegisterAppHandlerRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/ping", checkHealth(db))
}
