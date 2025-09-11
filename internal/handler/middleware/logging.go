package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"io"

	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type responseWriterWithBody struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriterWithBody) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID, err := uuid.NewRandom()

		if err != nil {
			logger.Log.Error(fmt.Errorf("UUID generation error: %w", err).Error())
		}

		var reqBody []byte

		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		headers := make(map[string]string, len(c.Request.Header))
		for key, vals := range c.Request.Header {
			headers[key] = strings.Join(vals, ",")
		}

		logger.Log.Info("Request received",
			zap.String("method", c.Request.Method),
			zap.String("URL", c.Request.URL.String()),
			zap.String("requestID", requestID.String()),
			zap.Any("headers", headers),
			zap.String("body", string(reqBody)),
		)

		customWriter := &responseWriterWithBody{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = customWriter
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		respHeaders := c.Writer.Header()
		respBody := customWriter.body.String()

		logger.Log.Info("Response sent",
			zap.Int("status", status),
			zap.String("statusText", http.StatusText(status)),
			zap.Duration("latency", latency),
			zap.Any("headers", respHeaders),
			zap.String("body", respBody),

			// если написать c.Writer.Size() будет другой результат, очень маленький(например 10 при не сжатых 180)
			zap.Int("size", customWriter.body.Len()),
		)
	}
}
