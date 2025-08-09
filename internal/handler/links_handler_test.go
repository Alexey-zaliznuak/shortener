package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			requestBody: "https://example.com",
			want: want{
				code:                201,
				responseContentType: "text/plain",
				notEmptyResponse:    true,
			},
		},
		{
			name:        "Create link with exists short URL",
			requestBody: "https://example.com",
			want: want{
				code:                201,
				responseContentType: "text/plain",
				notEmptyResponse:    true,
			},
		},
		{
			name:        "Create link with invalid URL",
			requestBody: "not valid link",
			want: want{
				code:                400,
				responseBody:        "create link error: invalid URL: 'not valid link'",
				responseContentType: "text/plain",
			},
		},
	}

	client := resty.New()

	router := NewRouter()
	cfg, _ := config.GetConfig(&config.FlagsInitialConfig{})

	RegisterLinksRoutes(router, &service.LinksService{AppConfig: cfg})

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

func Test_links_createLinkWithJsonAPI(t *testing.T) {
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
			requestBody: `{"url": "https://example.com"}`,
			want: want{
				code:                201,
				responseContentType: "text/plain",
				notEmptyResponse:    true,
			},
		},
		{
			name:        "Create link with exists short URL",
			requestBody: `{"url": "https://example.com"}`,
			want: want{
				code:                201,
				responseContentType: "text/plain",
				notEmptyResponse:    true,
			},
		},
		{
			name:        "Create link with invalid URL",
			requestBody: `{"url": "not valid link"}`,
			want: want{
				code:                400,
				responseBody:        "create link error: invalid URL: 'not valid link'",
				responseContentType: "text/plain",
			},
		},
	}

	client := resty.New()

	router := NewRouter()
	cfg, _ := config.GetConfig(&config.FlagsInitialConfig{})

	RegisterLinksRoutes(router, &service.LinksService{AppConfig: cfg})

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

	RegisterLinksRoutes(router, &service.LinksService{AppConfig: cfg})

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Get created link", func(t *testing.T) {
		fullURL := "https://example.com/"

		response, err := client.R().SetBody(fullURL).Post(server.URL)

		require.NoError(t, err)

		require.Equal(t, http.StatusCreated, response.StatusCode())

		shortcut := string(response.Body())

		response, err = client.R().Get(shortcut)

		require.NoError(t, err)

		assert.Equal(t, http.StatusTemporaryRedirect, response.StatusCode())
		assert.Equal(t, fullURL, response.Header().Get("Location"))
	})
}
