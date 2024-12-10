package repository

import (
	"errors"
	"sync"
)

type IURLRepository interface {
	Save(id string, originalURL string) error
	Find(id string) (string, error)
}

type URLRepository struct {
	store map[string]string // Хранилище для хранения пар ID и оригинальных URL
	mu    *sync.Mutex       // Мьютекс для обеспечения потокобезопасности
}

// NewStore создает новый экземпляр URLRepository.
func NewStore() *URLRepository {
	return &URLRepository{
		store: make(map[string]string),
		mu:    &sync.Mutex{},
	}
}

// Save сохраняет оригинальный URL по ID.
func (r *URLRepository) Save(id string, originalURL string) error {
	r.mu.Lock()         // Блокируем мьютекс
	defer r.mu.Unlock() // Разблокируем мьютекс после завершения работы

	r.store[id] = originalURL
	return nil
}

// Find ищет оригинальный URL по ID.
func (r *URLRepository) Find(id string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	originalURL, ok := r.store[id] // Проверяем, существует ли ID в хранилище
	if !ok {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}
