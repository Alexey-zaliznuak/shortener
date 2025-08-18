package handler

import (
	"github.com/Alexey-zaliznuak/shortener/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.Use(middleware.RequestAndResponseGzipCompressing())
	router.Use(middleware.RequestLogging())

	return router
}
