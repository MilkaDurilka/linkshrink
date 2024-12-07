package controller

import (
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
}

type URLController struct {
	service service.IURLService // Ссылка на сервис для работы с URL
	cfg     *config.Config
}

// NewURLController создает новый экземпляр URLController.
func NewURLController(cfg *config.Config, srv service.IURLService) *URLController {
	return &URLController{service: srv, cfg: cfg} // Возвращаем новый контроллер с заданным сервисом
}

// ShortenURL обрабатывает запрос на сокращение URL.
func (c *URLController) ShortenURL(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil || len(url) == 0 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	shortURL, err := c.service.Shorten(c.cfg.BaseURL, string(url))
	if err != nil {
		// Проверяем тип ошибки и отправляем соответствующий ответ
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		log.Println("Error shortening URL: %w", err)
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
	vars := mux.Vars(r)      // Получаем переменные маршрута
	id, exists := vars["id"] // Извлекаем ID из переменных маршрута

	if !exists {
		log.Println("Key 'id' not found in route variables")
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	originalURL, err := c.service.GetOriginalURL(id)

	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			http.Error(w, "URL not found", http.StatusBadRequest) // Если URL не найден, отправляем ответ 400 StatusBadRequest
			return
		}
		// Если другая ошибка, отправляем 500 Internal Server Error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
