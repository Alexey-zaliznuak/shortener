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
	shortUrl := c.Param("shortUrl")
	fullUrl, err := linksService.GetFullUrlFromShort(shortUrl)

	if err != nil || fullUrl == "" {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, fullUrl)
}

func createLink(c *gin.Context) {
	body, err := c.GetRawData()

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	link := &model.Link{FullUrl: string(body)}
	err = linksService.CreateLink(link)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.String(http.StatusCreated, "http://%s/%s", c.Request.Host, link.ShortUrl)
}

func init() {
	Router.POST("/", createLink)
	Router.GET("/:shortUrl", redirect)
}
