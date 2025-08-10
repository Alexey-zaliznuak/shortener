package handler

import (
	"github.com/Alexey-zaliznuak/shortener/internal/handler/middlewares"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.RequestLogging())

	return router
}
