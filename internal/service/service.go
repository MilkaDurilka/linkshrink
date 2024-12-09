package service

import (
	"errors"
	"fmt"
	"linkshrink/internal/repository"
	"log"
)

var (
	ErrInvalidURL  = errors.New("invalid URL")
	ErrURLNotFound = errors.New("URL not found")
)

type IURLService interface {
	Shorten(baseURL string, url string) (string, error)
	GetOriginalURL(id string) (string, error)
}

type URLService struct {
	repo repository.IURLRepository
}

func NewURLService(repo repository.IURLRepository) *URLService {
	return &URLService{ // Возвращаем новый сервис с заданным репозиторием
		repo: repo,
	}
}

// Shorten сокращает оригинальный URL.
func (s *URLService) Shorten(baseURL string, originalURL string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("url is empty: %w ", ErrInvalidURL)
	}

	id, err := s.repo.Save(originalURL)

	if err != nil {
		log.Println("Error saving URL:", err)
		return "", fmt.Errorf("failed to save URL with id %s: %w", id, err)
	}

	return baseURL + "/" + id, nil
}

// GetOriginalURL получает оригинальный URL по ID.
func (s *URLService) GetOriginalURL(id string) (string, error) {
	originalURL, err := s.repo.Find(id)
	if err != nil {
		return "", fmt.Errorf("%s not found  %w ", originalURL, ErrURLNotFound)
	}
	return originalURL, nil
}
