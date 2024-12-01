package controller

import (
	"errors"
	"io"
	"linkshrink/internal/service"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type IURLController interface {
	ShortenURL (w http.ResponseWriter, r *http.Request)
	RedirectURL(w http.ResponseWriter, r *http.Request)
}

type URLController struct {
    service service.IURLService // Ссылка на сервис для работы с URL
}

// NewURLController создает новый экземпляр URLController
func NewURLController(service service.IURLService) *URLController {
    return &URLController{service: service} // Возвращаем новый контроллер с заданным сервисом
}

// ShortenURL обрабатывает запрос на сокращение URL
func (c *URLController) ShortenURL(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
    if err != nil || len(url) == 0 {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }
		defer r.Body.Close() // Закрываем тело запроса после его чтения

    shortURL, err := c.service.Shorten(string(url))
		if err != nil {
			// Проверяем тип ошибки и отправляем соответствующий ответ
			if errors.Is(err, service.ErrInvalidURL) {
					http.Error(w, "Invalid URL", http.StatusBadRequest)
					return
			}
			http.Error(w, "Error shortening URL", http.StatusInternalServerError) 
			return
	}

    w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(shortURL))
}

// RedirectURL обрабатывает запрос на перенаправление по ID
func (c *URLController) RedirectURL(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r) // Получаем переменные маршрута
    id := vars["id"] // Извлекаем ID из переменных маршрута
    
    originalURL, err := c.service.GetOriginalURL(id)
		
    if err != nil {
        if errors.Is(err, service.ErrURLNotFound) {
            http.Error(w, "URL not found", http.StatusBadRequest) // Если URL не найден, отправляем ответ 400 StatusBadRequest
						return
        } 
        http.Error(w, "Internal server error", http.StatusInternalServerError) // Если другая ошибка, отправляем 500 Internal Server Error
        return
    }

    w.Header().Set("Location", originalURL)
    w.WriteHeader(http.StatusTemporaryRedirect)
}
