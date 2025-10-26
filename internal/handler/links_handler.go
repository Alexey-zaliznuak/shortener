package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/handler/audit"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func redirect(linksService *service.LinksService, authService *service.AuthService, auditor *audit.AuditorShortURLOperationManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortcut := c.Param("shortcut")
		fullURL, err := linksService.GetFullURLFromShort(shortcut)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		claims, err := authService.GetOrCreateAndSaveAuthorization(c)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		auditor.AuditNotify(audit.ShortURLActionGet, claims.ID, fullURL)

		if err != nil {
			if err == database.ErrObjectDeleted {
				c.Status(http.StatusGone)
				return
			}

			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, fullURL)
	}
}

func createLink(linksService *service.LinksService, auditor *audit.AuditorShortURLOperationManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		fullURL := string(body)

		link := &model.Link{FullURL: fullURL}

		link, claims, created, err := linksService.CreateLink(link.ToCreateDto(), c)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		auditor.AuditNotify(audit.ShortURLActionCreate, claims.ID, fullURL)

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

func createLinkWithJSONAPI(linksService *service.LinksService, auditor *audit.AuditorShortURLOperationManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		request := &model.CreateShortURLRequest{}
		err = json.Unmarshal(body, &request)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		l := &model.CreateLinkDto{FullURL: request.FullURL}

		link, claims, created, err := linksService.CreateLink(l, c)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		auditor.AuditNotify(audit.ShortURLActionCreate, claims.ID, request.FullURL)

		shortURL, err := linksService.BuildShortURL(link.Shortcut, c)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		status := http.StatusCreated
		if !created {
			status = http.StatusConflict
		}

		c.JSON(status, &model.CreateShortURLResponse{Result: shortURL})
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

		if err != nil && err != http.ErrNoCookie {
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

func deleteUserLinks(linksService *service.LinksService) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		request := make([]string, 0)

		err = json.Unmarshal(body, &request)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		err = linksService.DeleteUserLinks(request, c)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.Status(http.StatusAccepted)
	}
}

func RegisterLinksRoutes(router *gin.Engine, linksService *service.LinksService, authService *service.AuthService, auditor *audit.AuditorShortURLOperationManager, db *sql.DB) {
	router.GET("/:shortcut", redirect(linksService, authService, auditor))

	router.POST("/", createLink(linksService, auditor))
	router.POST("/api/shorten", createLinkWithJSONAPI(linksService, auditor))
	router.POST("/api/shorten/batch", createLinkBatch(linksService))

	router.GET("/api/user/urls", getUserLinks(linksService, authService))
	router.DELETE("/api/user/urls", deleteUserLinks(linksService))
}
