// Package handler provides HTTP handlers for the URL shortener service.
// It includes endpoints for creating, retrieving, and managing shortened URLs.
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

// redirect handles redirection from shortened URL to the original full URL.
// @Summary      Redirect to original URL
// @Description  Redirects from a shortened URL to the original full URL
// @Tags         links
// @Param        shortcut  path  string  true  "Short URL identifier"
// @Success      307  "Temporary redirect to the original URL"
// @Failure      400  {string}  string  "Invalid shortcut"
// @Failure      410  "Link has been deleted"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /{shortcut} [get]
func redirect(linksService *service.LinksService, authService *service.AuthService, auditor *audit.AuditorShortURLOperationManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortcut := c.Param("shortcut")
		fullURL, err := linksService.GetFullURLFromShort(shortcut)

		if err != nil {
			if err == database.ErrObjectDeleted {
				c.Status(http.StatusGone)
				return
			}

			c.String(http.StatusBadRequest, err.Error())
			return
		}

		claims, err := authService.GetOrCreateAndSaveAuthorization(c)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		auditor.AuditNotify(audit.ShortURLActionGet, claims.ID, fullURL)

		c.Redirect(http.StatusTemporaryRedirect, fullURL)
	}
}

// createLink creates a new shortened URL from a plain text body.
// @Summary      Create short URL (plain text)
// @Description  Creates a shortened URL from plain text body containing the full URL
// @Tags         links
// @Accept       plain
// @Produce      plain
// @Param        url  body  string  true  "Full URL to shorten (raw text)"
// @Success      201  {string}  string  "Created short URL"
// @Success      409  {string}  string  "URL already exists, returns existing short URL"
// @Failure      400  {string}  string  "Invalid request"
// @Router       / [post]
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

// createLinkWithJSONAPI creates a new shortened URL using JSON API format.
// @Summary      Create short URL (JSON)
// @Description  Creates a shortened URL using JSON request/response format
// @Tags         links
// @Accept       json
// @Produce      json
// @Param        request  body  model.CreateShortURLRequest  true  "URL to shorten"
// @Success      201  {object}  model.CreateShortURLResponse  "Short URL created"
// @Success      409  {object}  model.CreateShortURLResponse  "URL already exists"
// @Failure      400  {string}  string  "Invalid request"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /api/shorten [post]
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

// createLinkBatch creates multiple shortened URLs in a single request.
// @Summary      Create multiple short URLs
// @Description  Creates multiple shortened URLs in a single batch request with correlation IDs
// @Tags         links
// @Accept       json
// @Produce      json
// @Param        request  body  []model.CreateLinkWithCorrelationIDRequestItem  true  "Array of URLs to shorten with correlation IDs"
// @Success      201  {array}  model.CreateLinkWithCorrelationIDResponseItem  "Array of created short URLs"
// @Failure      400  {string}  string  "Invalid request"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /api/shorten/batch [post]
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

// getUserLinks retrieves all shortened URLs created by the current user.
// @Summary      Get user's URLs
// @Description  Retrieves all shortened URLs created by the authenticated user
// @Tags         user
// @Produce      json
// @Success      200  {array}  model.GetUserLinksRequestItem  "User's links"
// @Success      204  "User has no links or no valid authentication"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /api/user/urls [get]
// @Security     CookieAuth
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

// deleteUserLinks marks multiple shortened URLs as deleted for the current user.
// @Summary      Delete user's URLs
// @Description  Marks multiple shortened URLs as deleted (soft delete). Deletion is asynchronous.
// @Tags         user
// @Accept       json
// @Param        shortcuts  body  []string  true  "Array of shortcut identifiers to delete"
// @Success      202  "Deletion request accepted"
// @Failure      400  {string}  string  "Invalid request"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /api/user/urls [delete]
// @Security     CookieAuth
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

// RegisterLinksRoutes registers all URL shortener routes to the provided Gin engine.
// It sets up the following endpoints:
//   - GET /:shortcut - redirect to full URL
//   - POST / - create short URL (plain text)
//   - POST /api/shorten - create short URL (JSON)
//   - POST /api/shorten/batch - create multiple short URLs
//   - GET /api/user/urls - get all user's URLs
//   - DELETE /api/user/urls - delete user's URLs
//
// Parameters:
//   - router: Gin engine instance to register routes on
//   - linksService: service for managing links
//   - authService: service for user authentication
//   - auditor: auditor for logging URL operations
//   - db: database connection (currently unused, reserved for future use)
func RegisterLinksRoutes(router *gin.Engine, linksService *service.LinksService, authService *service.AuthService, auditor *audit.AuditorShortURLOperationManager, db *sql.DB) {
	router.GET("/:shortcut", redirect(linksService, authService, auditor))

	router.POST("/", createLink(linksService, auditor))
	router.POST("/api/shorten", createLinkWithJSONAPI(linksService, auditor))
	router.POST("/api/shorten/batch", createLinkBatch(linksService))

	router.GET("/api/user/urls", getUserLinks(linksService, authService))
	router.DELETE("/api/user/urls", deleteUserLinks(linksService))

	// router.GET("/api/public/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
