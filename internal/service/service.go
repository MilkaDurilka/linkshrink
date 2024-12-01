package service

import (
	"errors"
	"linkshrink/internal/repository"
)

var (
    ErrInvalidURL  = errors.New("invalid URL") 
    ErrURLNotFound = errors.New("URL not found")
)

type URLService struct {
    repo *repository.URLRepository
		idGenerator *IDGenerator
}

func NewURLService(repo *repository.URLRepository) *URLService {
  return &URLService{ // Возвращаем новый сервис с заданным репозиторием
			repo:        repo,
			idGenerator: NewIDGenerator(),
	} 
}

// Shorten сокращает оригинальный URL
func (s *URLService) Shorten(originalURL string) (string, error) {
    if originalURL == "" {
        return "", ErrInvalidURL
    }
	
    // Генерируем уникальный идентификатор
    id := s.idGenerator.GenerateID()
    if err := s.repo.Save(id, originalURL); err != nil {
        return "", err
    }
    return "http://localhost:8080/" + id, nil
}

// GetOriginalURL получает оригинальный URL по ID
func (s *URLService) GetOriginalURL(id string) (string, error) {
    originalURL, err := s.repo.Find(id)
    if err != nil {
        return "", ErrURLNotFound 
    }
    return originalURL, nil
}

