package controller

import (
	"encoding/json"
	"errors"
	"io"
	"linkshrink/internal/config"
	"linkshrink/internal/service"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type IURLController interface {
	ShortenURL(w http.ResponseWriter, r *http.Request)
	RedirectURL(w http.ResponseWriter, r *http.Request)
	ShortenURLJSON(w http.ResponseWriter, r *http.Request)
}

type URLController struct {
	service service.IURLService
	cfg     *config.Config
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

const (
	ErrInvalidURL = "Invalid URL"
)

// NewURLController создает новый экземпляр URLController.
func NewURLController(cfg *config.Config, srv service.IURLService) *URLController {
	return &URLController{service: srv, cfg: cfg}
}

// ShortenURL обрабатывает запрос на сокращение URL.
func (c *URLController) ShortenURL(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil || len(url) == 0 {
		http.Error(w, ErrInvalidURL, http.StatusBadRequest)
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	shortURL, err := c.service.Shorten(c.cfg.BaseURL, string(url))
	if err != nil {
		// Проверяем тип ошибки и отправляем соответствующий ответ.
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, ErrInvalidURL, http.StatusBadRequest)
			return
		}
		log.Println("Error shortening URL: %w", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	var data = []byte(shortURL)
	n, err := w.Write(data)
	if err != nil {
		log.Println("Error writing to the response stream: %w", err)
		return
	}

	if n != len(data) {
		log.Println("Error writing to the response stream: не все данные записаны")
		return
	}
}

// RedirectURL обрабатывает запрос на перенаправление по ID.
func (c *URLController) RedirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]

	if !ok {
		log.Println("Key 'id' not found in route variables")
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	originalURL, err := c.service.GetOriginalURL(id)

	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			http.Error(w, "URL not found", http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (c *URLController) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	// Декодируем JSON из тела запроса.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(req.URL) == 0 {
		http.Error(w, ErrInvalidURL, http.StatusBadRequest)
		return
	}

	// Вызываем метод контроллера для сокращения URL.
	shortURL, err := c.service.Shorten(c.cfg.BaseURL, req.URL)
	if err != nil {
		// Проверяем тип ошибки и отправляем соответствующий ответ.
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, ErrInvalidURL, http.StatusBadRequest)
			return
		}
		log.Println("Error shortening URL: %w", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Формируем ответ.
	resp := ShortenResponse{Result: shortURL}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
	}
}
