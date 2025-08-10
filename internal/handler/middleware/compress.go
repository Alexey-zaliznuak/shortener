package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var acceptedContentTypesForCompressing = strings.Join(([]string{"text/html", "application/json"}), "")

type responseWriterWithCompress struct {
	gin.ResponseWriter
	gzipWriter *gzip.Writer
}

func (w *responseWriterWithCompress) Write(b []byte) (int, error) {
	return w.gzipWriter.Write(b)
}

func (w *responseWriterWithCompress) WriteHeader(statusCode int) {
	w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWithCompress) Close() error {
	return w.gzipWriter.Close()
}

func newResponseWriterWithCompress(w gin.ResponseWriter) *responseWriterWithCompress {
	return &responseWriterWithCompress{
		ResponseWriter: w,
		gzipWriter:     gzip.NewWriter(w),
	}
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func RequestAndResponseGzipCompressing() gin.HandlerFunc {
	return func(c *gin.Context) {
		var newWriter *responseWriterWithCompress

		acceptEncodings := strings.Join(c.Request.Header.Values("Accept-Encoding"), "")
		contentType := c.Request.Header.Get("Content-Type")
		supportsGzip := strings.Contains(acceptedContentTypesForCompressing, contentType) && strings.Contains(acceptEncodings, "gzip")

		if supportsGzip {
			newWriter = newResponseWriterWithCompress(c.Writer)
			c.Writer = newWriter

			defer newWriter.Close()
		}

		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			compressReader, err := newCompressReader(c.Request.Body)
			if err != nil {
				newWriter.WriteHeader(http.StatusInternalServerError)
				newWriter.Write([]byte("Invalid content"))
				return
			}

			c.Request.Body = compressReader
			defer compressReader.Close()
		}

		c.Next()
	}
}
