package repository

import (
	"errors"
	"linkshrink/internal/config"
	"linkshrink/internal/utils/logger"

	"go.uber.org/zap"
)

var (
	ErrURLNotFound     = errors.New("URL not found")
	ErrIDAlreadyExists = errors.New("ID already exists")
)

type URLData struct {
	UUID        string `json:"uuid"`
	OriginalURL string `json:"original_url"`
}

type URLRepository interface {
	Save(originalURL string) (id string, err error)
	Find(id string) (string, error)
}

// NewStore создает новый экземпляр URLRepository.
func NewStore(cfg *config.Config, log logger.Logger) (URLRepository, error) {
	if cfg.DataBaseDSN != "" {
		dbLogger := log.With(zap.String("component", "DBStore"))
		return NewPostgresRepository(cfg.DataBaseDSN, dbLogger)
	}
	if cfg.FileStoragePath != "" {
		fileLogger := log.With(zap.String("component", "FileStore"))
		return NewFileStore(cfg.FileStoragePath, fileLogger)
	}
	memoryLogger := log.With(zap.String("component", "MemoryStore"))
	return NewMemoryStore(memoryLogger)
}
