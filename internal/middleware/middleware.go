package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func InitMiddlewares(logger *zap.Logger) func(http.Handler) http.Handler {
	return chain(
		gzipRequestMiddleware,
		gzipResponseMiddleware,
		loggingMiddleware(logger),
	)
}

func loggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем кастомный ResponseWriter для захвата кода статуса и размера ответа
			lrw := &LoggingResponseWriter{ResponseWriter: w}
			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			// Логируем информацию о запросе и ответе
			logger.Info("Request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", lrw.status),
				zap.Int64("size", lrw.size),
				zap.Duration("duration", duration),
			)
		})
	}
}

// LoggingResponseWriter - кастомный ResponseWriter для захвата статуса и размера ответа.
type LoggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.size += int64(n)

	if err != nil {
		return n, fmt.Errorf("failed to write response: %w", err)
	}

	return n, nil
}

// Middleware для обработки входящих сжатых запросов.
func gzipRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to create gzip reader", http.StatusBadRequest)
				return
			}

			defer func() {
				if err := reader.Close(); err != nil {
					http.Error(w, "Error closing reader:", http.StatusInternalServerError)
				}
			}()
			r.Body = io.NopCloser(reader)
		}
		next.ServeHTTP(w, r)
	})
}

const (
	ContentTypeHeader = "Content-Type"
)

// Middleware для сжатия ответов.
func gzipResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Encoding") == "gzip" && (r.Header.Get(ContentTypeHeader) == "application/json" || r.Header.Get(ContentTypeHeader) == "text/html") {
			var buf bytes.Buffer
			gzipWriter := gzip.NewWriter(&buf)

			defer func() {
				if err := gzipWriter.Close(); err != nil {
					http.Error(w, "Error closing gzipWriter:", http.StatusInternalServerError)
				}
			}()

			gzw := &gzipResponseWriter{ResponseWriter: w, Writer: gzipWriter}
			next.ServeHTTP(gzw, r)

			// gzipWriter.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set(ContentTypeHeader, gzw.Header().Get(ContentTypeHeader))
			w.WriteHeader(gzw.statusCode)
			_, err := buf.WriteTo(w)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Кастомный ResponseWriter для сжатия.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer     io.Writer
	statusCode int
}

func (g *gzipResponseWriter) Header() http.Header {
	return g.ResponseWriter.Header()
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	g.statusCode = code
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	n, err := g.Writer.Write(b)
	if err != nil {
		return n, fmt.Errorf("write gzip error : %w", err)
	}
	return n, nil
}

// Функция для объединения middleware.
func chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
