package repository

import (
	"errors"
	"sync"
)

type IURLRepository interface {
	Save(id, originalURL string) error
	Find(id string) (string, error)
}

type URLRepository struct {
	store map[string]string // Хранилище для хранения пар ID и оригинальных URL
	mu    sync.RWMutex      // Мьютекс для обеспечения потокобезопасности
}

// NewStore создает новый экземпляр URLRepository.
func NewStore() *URLRepository {
	return &URLRepository{store: make(map[string]string)}
}

// Save сохраняет оригинальный URL по ID.
func (r *URLRepository) Save(id, originalURL string) error {
	r.mu.Lock()         // Блокируем мьютекс для записи
	defer r.mu.Unlock() // Разблокируем мьютекс после завершения работы

	r.store[id] = originalURL
	return nil
}

// Find ищет оригинальный URL по ID.
func (r *URLRepository) Find(id string) (string, error) {
	r.mu.RLock()         // Блокируем мьютекс для чтения
	defer r.mu.RUnlock() // Разблокируем мьютекс после завершения работы

	originalURL, exists := r.store[id] // Проверяем, существует ли ID в хранилище
	if !exists {
		return "", errors.New("URL not found") // Если не найден, возвращаем ошибку
	}
	return originalURL, nil
}
