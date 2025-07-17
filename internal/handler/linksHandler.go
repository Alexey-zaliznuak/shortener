package handler

import (
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

var linksService = service.NewLinksService(database.Client)

func redirect(c *gin.Context) {
	shortcut := c.Param("shortcut")
	fullURL, err := linksService.GetFullURLFromShort(shortcut)

	if err != nil || fullURL == "" {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, fullURL)
}

func createLink(c *gin.Context) {
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

func init() {
	Router.POST("/", createLink)
	Router.GET("/:shortcut", redirect)
}
