package handler

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/link"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRandomString() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	result := make([]rune, 10)

	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
func generateRandomURL() string {
	return fmt.Sprintf("https://example.com/%s", generateRandomString())
}

func Test_links_createLink(t *testing.T) {
	type want struct {
		code                int
		responseBody        string
		notEmptyResponse    bool
		responseContentType string
	}
	type test struct {
		name        string
		requestURL  string
		requestBody string
		want        want
	}

	createLinkTests := []test{
		{
			name:        "Create valid link",
			requestBody: generateRandomURL(),
			want: want{
				code:                http.StatusCreated,
				responseContentType: "text/plain",
				notEmptyResponse:    true,
			},
		},
		{
			name:        "Create link with exists short URL",
			requestBody: generateRandomURL(),
			want: want{
				code:                http.StatusCreated,
				responseContentType: "text/plain",
				notEmptyResponse:    true,
			},
		},
		{
			name:        "Create link with invalid URL",
			requestBody: "not valid link",
			want: want{
				code:                http.StatusBadRequest,
				responseBody:        "create link error: invalid URL: 'not valid link'",
				responseContentType: "text/plain",
			},
		},
	}

	client := resty.New()
	router := NewRouter()

	cfg, _ := config.GetConfig(&config.FlagsInitialConfig{})
	var db *sql.DB
	var err error

	if cfg.DB.DatabaseDSN != "" {
		fmt.Println("DATABASE DSN")
		fmt.Println(cfg.DB.DatabaseDSN)
		db, err = database.NewDatabaseConnectionPool(cfg)
		require.NoError(t, err)
	}

	r, err := link.NewLinksRepository(context.Background(), cfg, db)
	require.NoError(t, err)

	RegisterLinksRoutes(router, service.NewLinksService(r, cfg), db)

	server := httptest.NewServer(router)
	defer server.Close()

	for _, test := range createLinkTests {
		t.Run(test.name, func(t *testing.T) {
			response, err := client.R().SetBody(test.requestBody).Post(server.URL + test.requestURL)

			require.NoError(t, err)

			t.Logf("body: '%s'", string(response.Body()))

			assert.Equal(t, test.want.code, response.StatusCode())

			if test.want.responseBody != "" {
				assert.Equal(t, test.want.responseBody, string(response.Body()))
				assert.Contains(t, response.Header().Get("Content-Type"), test.want.responseContentType)
			}

			if test.want.notEmptyResponse {
				assert.NotEmpty(t, string(response.Body()))
			}
		})
	}
}

func Test_links_createLinkWithJSONAPI(t *testing.T) {
	type want struct {
		code                int
		responseBody        string
		notEmptyResponse    bool
		responseContentType string
	}
	type test struct {
		name        string
		requestBody string
		want        want
	}

	createLinkTests := []test{
		{
			name:        "Create valid link",
			requestBody: fmt.Sprintf(`{"url": "%s"}`, generateRandomURL()),
			want: want{
				code:                http.StatusCreated,
				responseContentType: "text/plain",
				notEmptyResponse:    true,
			},
		},
		{
			name:        "Create link with exists short URL",
			requestBody: fmt.Sprintf(`{"url": "%s"}`, generateRandomURL()),
			want: want{
				code:                http.StatusCreated,
				responseContentType: "text/plain",
				notEmptyResponse:    true,
			},
		},
		{
			name:        "Create link with invalid URL",
			requestBody: `{"url": "not valid link"}`,
			want: want{
				code:                http.StatusBadRequest,
				responseBody:        "create link error: invalid URL: 'not valid link'",
				responseContentType: "text/plain",
			},
		},
	}

	client := resty.New()
	router := NewRouter()

	cfg, _ := config.GetConfig(&config.FlagsInitialConfig{})
	var db *sql.DB
	var err error

	if cfg.DB.DatabaseDSN != "" {
		db, err = database.NewDatabaseConnectionPool(cfg)
		require.NoError(t, err)
	}

	r, err := link.NewLinksRepository(context.Background(), cfg, db)
	require.NoError(t, err)

	RegisterLinksRoutes(router, service.NewLinksService(r, cfg), db)

	server := httptest.NewServer(router)
	defer server.Close()

	for _, test := range createLinkTests {
		t.Run(test.name, func(t *testing.T) {
			var response *resty.Response
			var err error

			requestURL := server.URL + "/api/shorten"
			response, err = client.R().SetBody(test.requestBody).Post(requestURL)

			require.NoError(t, err)

			t.Logf("body: '%s'", string(response.Body()))

			assert.Equal(t, test.want.code, response.StatusCode())

			if test.want.responseBody != "" {
				assert.Equal(t, test.want.responseBody, string(response.Body()))
				assert.Contains(t, response.Header().Get("Content-Type"), test.want.responseContentType)
			}

			if test.want.notEmptyResponse {
				assert.NotEmpty(t, string(response.Body()))
			}
		})
	}
}

func Test_links_CreateAndGet(t *testing.T) {
	client := resty.New().SetRedirectPolicy(resty.RedirectPolicyFunc(
		func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	))

	router := NewRouter()

	cfg, _ := config.GetConfig(&config.FlagsInitialConfig{})
	var db *sql.DB
	var err error

	if cfg.DB.DatabaseDSN != "" {
		db, err = database.NewDatabaseConnectionPool(cfg)
		require.NoError(t, err)
	}

	r, err := link.NewLinksRepository(context.Background(), cfg, db)
	require.NoError(t, err)

	RegisterLinksRoutes(router, service.NewLinksService(r, cfg), db)

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Get created link", func(t *testing.T) {
		fullURL := generateRandomURL()

		response, err := client.R().SetBody(fullURL).Post(server.URL)

		require.NoError(t, err)

		require.Equal(t, http.StatusCreated, response.StatusCode())

		res := strings.Split(string(response.Body()), "/")

		shortcut := res[len(res)-1]
		response, err = client.R().Get(server.URL + "/" + shortcut)

		require.NoError(t, err)

		assert.Equal(t, http.StatusTemporaryRedirect, response.StatusCode())
		assert.Equal(t, fullURL, response.Header().Get("Location"))
	})
}
