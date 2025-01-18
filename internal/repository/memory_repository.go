package repository

import (
	"linkshrink/internal/utils"
	"linkshrink/internal/utils/logger"
	"sync"
)

type MemoryStore struct {
	Store       map[string]string // Хранилище для хранения пар ID и оригинальных URL
	mu          *sync.Mutex       // Мьютекс для обеспечения потокобезопасности
	logger      logger.Logger
	idGenerator *utils.IDGenerator
}

// NewMemoryStore создает новый экземпляр MemoryStore.
func NewMemoryStore(log logger.Logger) (*MemoryStore, error) {
	repo := &MemoryStore{
		Store:       make(map[string]string),
		mu:          &sync.Mutex{},
		logger:      log,
		idGenerator: utils.NewIDGenerator(),
	}

	return repo, nil
}

// Save сохраняет оригинальный URL по ID.
func (r *MemoryStore) Save(originalURL string) (string, error) {
	id := r.idGenerator.GenerateID()

	r.mu.Lock() // Блокируем мьютекс
	defer r.mu.Unlock()

	_, ok := r.Store[id]
	if !ok {
		r.Store[id] = originalURL
		return id, nil
	}

	return "", ErrIDAlreadyExists
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
