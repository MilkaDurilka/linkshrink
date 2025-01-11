package repository

import (
	"errors"
	filestore "linkshrink/internal/repository/file_store"
	memorystore "linkshrink/internal/repository/memory_store"
	"linkshrink/internal/utils/logger"
)

var (
	ErrURLNotFound     = errors.New("URL not found")
	ErrIDAlreadyExists = errors.New("ID already exists")
)

type URLData struct {
	UUID        string `json:"uuid"`
	OriginalURL string `json:"original_url"`
}

type IURLRepository interface {
	Save(id string, originalURL string) error
	Find(id string) (string, error)
}

type URLRepository struct {
	Store map[string]string // Хранилище для хранения пар ID и оригинальных URL
}

// NewStore создает новый экземпляр URLRepository.
func NewStore(storeType string, filePath string, log logger.Logger) IURLRepository {
	if storeType == "file" {
		return filestore.NewFileStore(filePath, log)
	}
	return memorystore.NewMemoryStore(log)
}
