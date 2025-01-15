package repository

import (
	"linkshrink/internal/utils/logger"
	"sync"
)

type IMemoryStore interface {
	Save(id string, originalURL string) error
	Find(id string) (string, error)
}

type MemoryStore struct {
	Store  map[string]string // Хранилище для хранения пар ID и оригинальных URL
	mu     *sync.Mutex       // Мьютекс для обеспечения потокобезопасности
	logger logger.Logger
}

// NewMemoryStore создает новый экземпляр MemoryStore.
func NewMemoryStore(log logger.Logger) (*MemoryStore, error) {
	repo := &MemoryStore{
		Store:  make(map[string]string),
		mu:     &sync.Mutex{},
		logger: log,
	}

	return repo, nil
}

// Save сохраняет оригинальный URL по ID.
func (r *MemoryStore) Save(id string, originalURL string) error {
	r.mu.Lock() // Блокируем мьютекс
	defer r.mu.Unlock()

	_, ok := r.Store[id]
	if !ok {
		r.Store[id] = originalURL
		return nil
	}

	return ErrIDAlreadyExists
}

// Find ищет оригинальный URL по ID.
func (r *MemoryStore) Find(id string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	originalURL, ok := r.Store[id] // Проверяем, существует ли ID в хранилище
	if !ok {
		return "", ErrURLNotFound
	}
	return originalURL, nil
}
