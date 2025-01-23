package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"linkshrink/internal/config"
	"linkshrink/internal/service"
	"linkshrink/internal/utils"
	errorsUtils "linkshrink/internal/utils/errors"
	"linkshrink/internal/utils/logger"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type URLController interface {
	ShortenURL(ctx context.Context, w http.ResponseWriter, r *http.Request)
	RedirectURL(w http.ResponseWriter, r *http.Request)
	ShortenURLJSON(ctx context.Context, w http.ResponseWriter, r *http.Request)
	BatchShortenURL(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

type URLControllerImpl struct {
	service service.IURLService
	cfg     *config.Config
	logger  logger.Logger
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

const (
	ErrInvalidURL = "Invalid URL"
	ErrInternal   = "Internal server error"
)

// NewURLController создает новый экземпляр URLController.
func NewURLController(
	cfg *config.Config,
	srv service.IURLService,
	log logger.Logger,
) *URLControllerImpl {
	return &URLControllerImpl{service: srv, cfg: cfg, logger: log}
}

// ShortenURL обрабатывает запрос на сокращение URL.
func (c *URLControllerImpl) ShortenURL(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil || len(url) == 0 {
		http.Error(w, ErrInvalidURL, http.StatusBadRequest)
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			c.logger.Error("Error closing response body", zap.Error(err))
		}
	}()

	shortURL, err := c.service.Shorten(ctx, c.cfg.BaseURL, string(url))
	statusCode := http.StatusCreated
	if err != nil {
		// Проверяем тип ошибки и отправляем соответствующий ответ.
		if errorsUtils.IsUniqueViolation(err) {
			statusCode = http.StatusConflict
		} else {
			if errors.Is(err, service.ErrInvalidURL) {
				http.Error(w, ErrInvalidURL, http.StatusBadRequest)
				return
			}

			c.logger.Error("Error shortening URL", zap.Error(err))
			http.Error(w, ErrInternal, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "text/plain")
	var data = []byte(shortURL)
	n, err := w.Write(data)
	if err != nil {
		c.logger.Error("Error writing to the response stream", zap.Error(err))
		return
	}

	if n != len(data) {
		c.logger.Error("Error writing to the response stream: не все данные записаны")
		return
	}
}

// RedirectURL обрабатывает запрос на перенаправление по ID.
func (c *URLControllerImpl) RedirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]

	if !ok {
		c.logger.Error("Key 'id' not found in route variables")
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	originalURL, err := c.service.GetOriginalURL(id)

	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			http.Error(w, "URL not found", http.StatusBadRequest)
			return
		}

		c.logger.Error("Error on GetOriginalURL", zap.Error(err))
		http.Error(w, ErrInternal, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	_, err = w.Write([]byte(originalURL))
	if err != nil {
		c.logger.Error("Error on Write", zap.Error(err))
		http.Error(w, ErrInternal, http.StatusInternalServerError)
		return
	}
}

func (c *URLControllerImpl) ShortenURLJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	// Декодируем JSON из тела запроса.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Error("Error on decoding", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(req.URL) == 0 {
		http.Error(w, ErrInvalidURL, http.StatusBadRequest)
		return
	}

	// Вызываем метод контроллера для сокращения URL.
	shortURL, err := c.service.Shorten(ctx, c.cfg.BaseURL, req.URL)

	statusCode := http.StatusCreated
	if err != nil {
		// Проверяем тип ошибки и отправляем соответствующий ответ.
		if errorsUtils.IsUniqueViolation(err) {
			statusCode = http.StatusConflict
		} else {
			// Проверяем тип ошибки и отправляем соответствующий ответ.
			if errors.Is(err, service.ErrInvalidURL) {
				http.Error(w, ErrInvalidURL, http.StatusBadRequest)
				return
			}
			c.logger.Error("Error shortening URL", zap.Error(err))
			http.Error(w, ErrInternal, http.StatusInternalServerError)
			return
		}
	}

	// Формируем ответ.
	resp := ShortenResponse{Result: shortURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.logger.Error("Error on encoding", zap.Error(err))
		http.Error(w, ErrInternal, http.StatusInternalServerError)
	}
}

func (c *URLControllerImpl) BatchShortenURL(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var requests []utils.BatchShortenParam

	// Декодируем JSON из тела запроса.
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		c.logger.Error("Error decoding request payload", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(requests) == 0 {
		http.Error(w, "Empty batch not allowed", http.StatusBadRequest)
		return
	}

	responses, err := c.service.BatchShorten(ctx, c.cfg.BaseURL, requests)

	if err != nil {
		c.logger.Error("Error batch shortening URL", zap.Error(err))
		http.Error(w, ErrInternal, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		c.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, ErrInternal, http.StatusInternalServerError)
	}
}
