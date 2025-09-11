package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository"
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
		link, created, err := linksService.CreateLink(link.ToCreateDto(), c)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		url, err := linksService.BuildShortURL(link.Shortcut, c)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		status := http.StatusCreated
		if !created {
			fmt.Printf("Duplicate: %s", link.FullURL)
			status = http.StatusConflict
		}

		c.String(status, url)
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

		l := &model.CreateLinkDto{FullURL: request.FullURL}
		link, created, err := linksService.CreateLink(l, c)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		shortURL, err := linksService.BuildShortURL(link.Shortcut, c)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		status := http.StatusCreated
		if !created {
			status = http.StatusConflict
		}

		c.JSON(status, &createShortURLResponse{Result: shortURL})
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
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

func getUserLinks(linksService *service.LinksService, authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := http.StatusOK
		links, err := linksService.GetUserLinks(c)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		_, err = authService.CreateAndSaveAuthorization(c)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		if err == http.ErrNoCookie {
			c.String(http.StatusNoContent, "")

			_, err = authService.CreateAndSaveAuthorization(c)

			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			return
		}

		if err == repository.ErrTokenValidation {
			_, err = authService.CreateAndSaveAuthorization(c)
			status = http.StatusNoContent
		}

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		if len(links) == 0 {
			status = http.StatusNoContent
		}

		c.JSON(status, links)
	}
}

func RegisterLinksRoutes(router *gin.Engine, linksService *service.LinksService, authService *service.AuthService, db *sql.DB) {
	router.GET("/:shortcut", redirect(linksService))

	router.POST("/", createLink(linksService))
	router.POST("/api/shorten", createLinkWithJSONAPI(linksService))
	router.POST("/api/shorten/batch", createLinkBatch(linksService))

	router.GET("/api/user/urls", getUserLinks(linksService, authService))
}
