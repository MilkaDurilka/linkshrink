package repository

import (
	"errors"
	"linkshrink/internal/config"
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
func NewStore(cfg *config.Config, log logger.Logger) (IURLRepository, error) {
	if cfg.DataBaseDSN != "" {
		return NewPostgresRepository(cfg.DataBaseDSN, log)
	}
	if cfg.FileStoragePath != "" {
		return NewFileStore(cfg.FileStoragePath, log)
	}
	return NewMemoryStore(log)
}
