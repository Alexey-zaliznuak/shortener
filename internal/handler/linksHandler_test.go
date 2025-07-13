package handler

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_defaultHandler_createLink(t *testing.T) {
	type want struct {
		request             string
		code                int
		response            string
		notEmptyResponse    bool
		responseContentType string
	}
	type test struct {
		name string
		want want
	}

	createLinkTests := []test{
		{
			name: "Create valid link",
			want: want{
				request:             "https://google.com",
				code:                201,
				responseContentType: "text/plain",
				notEmptyResponse: true,
			},
		},
		{
			name: "Create link with exists short url",
			want: want{
				request:             "https://google.com",
				code:                201,
				responseContentType: "text/plain",
				notEmptyResponse: true,
			},
		},
		{
			name: "Create link with invalid url",
			want: want{
				request:             "1234",
				code:                400,
				response:            "create link error: invalid url: '1234'\n", // \n made by http.Error
				responseContentType: "text/plain",
			},
		},
	}

	for _, test := range createLinkTests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/", strings.NewReader(test.want.request))

			w := httptest.NewRecorder()
			defaultHandler(w, request)

			response := w.Result()

			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			t.Logf("body: '%s'", string(body))

			if test.want.response != "" {
				assert.Equal(t, test.want.response, string(body))
				assert.Contains(t, response.Header.Get("Content-Type"), test.want.responseContentType)
			}

			if test.want.notEmptyResponse {
				assert.NotEmpty(t, string(body))
			}
			assert.Equal(t, test.want.code, response.StatusCode)
		})
	}
}
