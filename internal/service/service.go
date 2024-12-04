package service

import (
	"errors"
	"linkshrink/internal/repository"
)

var (
	ErrInvalidURL  = errors.New("invalid URL")
	ErrURLNotFound = errors.New("URL not found")
)

// Определите интерфейс URLService.
type IURLService interface {
	Shorten(baseURL string, url string) (string, error)
	GetOriginalURL(id string) (string, error)
}

type URLService struct {
	repo        repository.IURLRepository
	idGenerator *IDGenerator
}

func NewURLService(repo repository.IURLRepository) *URLService {
	return &URLService{ // Возвращаем новый сервис с заданным репозиторием
		repo:        repo,
		idGenerator: NewIDGenerator(),
	}
}

// Shorten сокращает оригинальный URL.
func (s *URLService) Shorten(baseURL string, originalURL string) (string, error) {
	if originalURL == "" {
		return "", ErrInvalidURL
	}

	// Генерируем уникальный идентификатор
	id := s.idGenerator.GenerateID()
	if err := s.repo.Save(id, originalURL); err != nil {
		return "", err
	}

	return baseURL + "/" + id, nil
}

// GetOriginalURL получает оригинальный URL по ID.
func (s *URLService) GetOriginalURL(id string) (string, error) {
	originalURL, err := s.repo.Find(id)
	if err != nil {
		return "", ErrURLNotFound
	}
	return originalURL, nil
}
