package repository

import (
	"context"
	"errors"
	"linkshrink/internal/config"
	"linkshrink/internal/utils/logger"
	"linkshrink/internal/utils/transaction"

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
	Save(ctx context.Context, originalURL string) (id string, err error)
	Find(id string) (string, error)
}

// NewStore создает новый экземпляр URLRepository.
func NewStore(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
) (URLRepository, transaction.TxManager, error) {
	if cfg.DataBaseDSN != "" {
		dbLogger := log.With(zap.String("component", "DBStore"))
		pgRep, err := NewPostgresRepository(ctx, cfg.DataBaseDSN, dbLogger)
		txManager := transaction.NewTransactionManager(pgRep.db)
		return pgRep, txManager, err
	}
	if cfg.FileStoragePath != "" {
		fileLogger := log.With(zap.String("component", "FileStore"))
		fileRep, err := NewFileStore(cfg.FileStoragePath, fileLogger)
		return fileRep, nil, err
	}
	memoryLogger := log.With(zap.String("component", "MemoryStore"))
	memoryRep, err := NewMemoryStore(memoryLogger)
	return memoryRep, nil, err
}
