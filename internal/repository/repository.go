package repository

import (
	"errors"
	"linkshrink/internal/utils"
	"sync"
)

type IURLRepository interface {
	Save(originalURL string) (string, error)
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
func (r *URLRepository) Save(originalURL string) (string, error) {
	r.mu.Lock()         // Блокируем мьютекс для записи
	defer r.mu.Unlock() // Разблокируем мьютекс после завершения работы

	const maxAttempts = 10 // Максимальное количество попыток
	var id string
	attempts := 0

	for attempts < maxAttempts {
		id = utils.NewIDGenerator().GenerateID()
		if _, ok := r.store[id]; !ok {
			r.store[id] = originalURL
			return id, nil
		}
		attempts++
	}

	return "", errors.New("internal Server Error")
}

// Find ищет оригинальный URL по ID.
func (r *URLRepository) Find(id string) (string, error) {
	r.mu.Lock()         // Блокируем мьютекс для чтения
	defer r.mu.Unlock() // Разблокируем мьютекс после завершения работы

	originalURL, ok := r.store[id] // Проверяем, существует ли ID в хранилище
	if !ok {
		return "", errors.New("URL not found") // Если не найден, возвращаем ошибку
	}
	return originalURL, nil
}
