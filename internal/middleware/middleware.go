package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"linkshrink/internal/utils/logger"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func InitMiddlewares(log logger.Logger) func(http.Handler) http.Handler {
	return chain(
		GzipRequestMiddleware(log),
		GzipResponseMiddleware(log),
		loggingMiddleware(log),
	)
}

func loggingMiddleware(log logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		componentLogger := log.With(zap.String("component", "loggingMiddleware"))

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем кастомный ResponseWriter для захвата кода статуса и размера ответа
			lrw := &LoggingResponseWriter{ResponseWriter: w}
			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			// Логируем информацию о запросе и ответе
			componentLogger.Info("Request",
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

// GzipRequestMiddleware для обработки входящих сжатых запросов.
func GzipRequestMiddleware(log logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		componentLogger := log.With(zap.String("component", "GzipRequestMiddleware"))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {
				reader, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Failed to create gzip reader", http.StatusBadRequest)
					return
				}

				defer func() {
					if err := reader.Close(); err != nil {
						componentLogger.Error("Error closing reader", zap.Error(err))
						http.Error(w, "Internal server error", http.StatusInternalServerError)
					}
				}()

				body, err := io.ReadAll(reader)
				if err != nil {
					http.Error(w, "Failed to read body", http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(body))
				r.Header.Set("Content-Encoding", "")
			}
			next.ServeHTTP(w, r)
		})
	}
}

const (
	ContentTypeHeader = "Content-Type"
)

// GzipResponseMiddleware для сжатия ответов.
func GzipResponseMiddleware(log logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		componentLogger := log.With(zap.String("component", "GzipResponseMiddleware"))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем наличие "gzip" в заголовке Accept-Encoding
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") &&
				(r.Header.Get(ContentTypeHeader) == "application/json" || r.Header.Get(ContentTypeHeader) == "text/html") {
				// Проверяем размер ответа, чтобы избежать сжатия маленьких файлов
				gzw := &gzipResponseWriter{ResponseWriter: w, Writer: nil}
				next.ServeHTTP(gzw, r)

				const minGzipSize = 1400

				if gzw.size < minGzipSize { // Если размер меньше 1400 байт, не сжимаем
					// Устанавливаем заголовки и возвращаем ответ без сжатия
					w.Header().Set(ContentTypeHeader, gzw.Header().Get(ContentTypeHeader))
					w.WriteHeader(gzw.statusCode)
					// Просто возвращаем оригинальный ответ
					return
				}

				// Создаем gzip.Writer и используем его
				var buf bytes.Buffer
				gzipWriter := gzip.NewWriter(&buf)

				defer func() {
					if err := gzipWriter.Close(); err != nil {
						componentLogger.Error("Error closing gzipWriter", zap.Error(err))
						http.Error(w, "Internal server error", http.StatusInternalServerError)
					}
				}()

				// Сбрасываем gzip.Writer для повторного использования
				gzipWriter.Reset(&buf)
				gzw.Writer = gzipWriter

				// Обрабатываем запрос снова с gzipWriter
				next.ServeHTTP(gzw, r)

				// Устанавливаем заголовки для gzip
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set(ContentTypeHeader, gzw.Header().Get(ContentTypeHeader))
				w.WriteHeader(gzw.statusCode)

				// Записываем сжатый ответ
				_, err := buf.WriteTo(w)
				if err != nil {
					componentLogger.Error("Error writing compressed response", zap.Error(err))
					return
				}
				return
			}
			// Если gzip не поддерживается, просто обрабатываем запрос
			next.ServeHTTP(w, r)
		})
	}
}

// Кастомный ResponseWriter для сжатия.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer     io.Writer
	statusCode int
	size       int64
}

func (g *gzipResponseWriter) Header() http.Header {
	return g.ResponseWriter.Header()
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	g.statusCode = code
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Writer == nil {
		n, err := g.ResponseWriter.Write(b)
		if err != nil {
			return n, fmt.Errorf("failed to write response: %w", err)
		}
		return n, nil
	}
	n, err := g.Writer.Write(b)
	g.size += int64(n) //
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
