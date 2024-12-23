package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
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
