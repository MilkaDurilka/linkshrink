package service

import (
	"errors"
	"fmt"
	"linkshrink/internal/repository"
)

var (
	ErrInvalidURL     = errors.New("invalid URL")
	ErrURLNotFound    = errors.New("URL not found")
	ErrInternalServer = errors.New("internal Server Error")
)

type IURLService interface {
	Shorten(baseURL string, url string) (string, error)
	GetOriginalURL(id string) (string, error)
	BeginTransaction() (repository.Transaction, error)
}

type URLService struct {
	repo        repository.URLRepository
	idGenerator *IDGenerator
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{ // Возвращаем новый сервис с заданным репозиторием
		repo:        repo,
		idGenerator: NewIDGenerator(),
	}
}

// Shorten сокращает оригинальный URL.
func (s *URLService) Shorten(baseURL string, originalURL string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("url is empty: %w ", ErrInvalidURL)
	}

	const maxAttempts = 10 // Максимальное количество попыток
	attempts := 0

	for attempts < maxAttempts {
		id := s.idGenerator.GenerateID()
		err := s.repo.Save(id, originalURL)

		if err == nil {
			return baseURL + "/" + id, nil
		}
		attempts++
	}

	return "", fmt.Errorf("%w: number of attempts exceeded: %s", ErrInternalServer, originalURL)
}

// GetOriginalURL получает оригинальный URL по ID.
func (s *URLService) GetOriginalURL(id string) (string, error) {
	originalURL, err := s.repo.Find(id)
	if err != nil {
		return "", fmt.Errorf("%s not found  %w ", originalURL, ErrURLNotFound)
	}
	return originalURL, nil
}

func (s *URLService) BeginTransaction() (repository.Transaction, error) {
	var transactionRepo repository.TransactableRepository

	if postgresRepo, ok := s.repo.(repository.TransactableRepository); ok {
		transactionRepo = postgresRepo
	} else {
		return nil, fmt.Errorf("transactionRepo  does not implement ITransactableRepository: %w ", ErrInvalidURL)
	}

	tx, err := transactionRepo.Begin()
	if err != nil {
		return nil, fmt.Errorf("tx begin error: %w ", ErrInvalidURL)
	}

	return tx, nil
}
