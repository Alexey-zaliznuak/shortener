package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

type createShortURLRequest struct {
	FullURL string `json:"url"`
}
type createShortURLResponse struct {
	Result string `json:"result"`
}

func redirect(linksService *service.LinksService) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortcut := c.Param("shortcut")
		fullURL, err := linksService.GetFullURLFromShort(shortcut)

		if err != nil {
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

func createLinkWithJSONAPI(linksService *service.LinksService) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		request := &createShortURLRequest{}
		err = json.Unmarshal(body, &request)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		link := &model.Link{FullURL: request.FullURL}
		err = linksService.CreateLink(link)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		shortURL, err := linksService.BuildShortURL(link.Shortcut, c)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusCreated, &createShortURLResponse{Result: shortURL})
	}
}

func createLinkBatch(linksService *service.LinksService) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		request := make([]*model.CreateLinkWithCorrelationIDRequestItem, 0, 100)

		err = json.Unmarshal(body, &request)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		response, err := linksService.BulkCreateWithCorrelationID(request, c)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

func RegisterLinksRoutes(router *gin.Engine, linksService *service.LinksService, db *sql.DB) {
	router.GET("/:shortcut", redirect(linksService))

	router.POST("/", createLink(linksService))
	router.POST("/api/shorten", createLinkWithJSONAPI(linksService))
	router.POST("/api/shorten/batch", createLinkBatch(linksService))
}
