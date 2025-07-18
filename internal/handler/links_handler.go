package handler

import (
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func redirect(linksService *service.LinksService) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortcut := c.Param("shortcut")
		fullURL, err := linksService.GetFullURLFromShort(shortcut)

		if err != nil || fullURL == "" {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, fullURL)
	}
}

func createLink(linksService *service.LinksService) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		link := &model.Link{FullURL: string(body)}
		err = linksService.CreateLink(link)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		url, err := linksService.BuildShortURL(link.Shortcut, c)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.String(http.StatusCreated, url)
	}
}

func SetupLinksRoutes(router *gin.Engine, linksService *service.LinksService) {
	router.POST("/", createLink(linksService))
	router.GET("/:shortcut", redirect(linksService))
}
