package service

import (
	"errors"
	"fmt"
	"linkshrink/internal/repository"
	"linkshrink/internal/utils"
	errorsUtils "linkshrink/internal/utils/errors"
)

var (
	ErrInvalidURL     = errors.New("invalid URL")
	ErrURLNotFound    = errors.New("URL not found")
	ErrInternalServer = errors.New("internal Server Error")
)

type IURLService interface {
	Shorten(baseURL string, url string) (string, error)
	GetOriginalURL(id string) (string, error)
	BatchShorten(baseURL string, params []utils.BatchShortenParam) ([]utils.BatchShortenReturnParam, error)
	// BeginTransaction() (repository.Transaction, error)
}

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{ // Возвращаем новый сервис с заданным репозиторием
		repo: repo,
	}
}

// Shorten сокращает оригинальный URL.
func (s *URLService) BatchShorten(baseURL string, params []utils.BatchShortenParam) ([]utils.BatchShortenReturnParam, error) {
	var res []utils.BatchShortenReturnParam
	for _, param := range params {
		if len(param.OriginalURL) == 0 {
			return nil, fmt.Errorf("url %s is empty: %w ", param.CorrelationID, ErrInvalidURL)
		}
	}

	const maxAttempts = 10 // Максимальное количество попыток
	attempts := 0

	for attempts < maxAttempts {

	data, err := s.repo.SaveAll(params)

	for i, item := range data {
		res[i] = utils.BatchShortenReturnParam{
			CorrelationID: item.CorrelationID,
			ShortURL: baseURL + "/" + item.ID,
		}
	}

	if err != nil {
		if errorsUtils.IsUniqueViolation(err) {
			return res, fmt.Errorf("some url is not unique: %w ", err)
		}
		return nil, fmt.Errorf("batch save error ", err)
	}
		if err == nil {
			return res, nil //baseURL + "/" + id
		}
		attempts++
	}

	return nil, fmt.Errorf("%w: number of attempts exceeded", ErrInternalServer)

	// if originalURL == "" {
	// 	return "", fmt.Errorf("url is empty: %w ", ErrInvalidURL)
	// }

	// const maxAttempts = 10 // Максимальное количество попыток
	// attempts := 0

	// for attempts < maxAttempts {
	// 	id, err := s.repo.Save(originalURL)

	// 	if errorsUtils.IsUniqueViolation(err) {
	// 		return baseURL + "/" + id, fmt.Errorf("url %s is not unique: %w ", originalURL, err)
	// 	}

	// 	if err == nil {
	// 		return baseURL + "/" + id, nil
	// 	}
	// 	attempts++
	// }

	// return "", fmt.Errorf("%w: number of attempts exceeded: %s", ErrInternalServer, originalURL)
}

func (s *URLService) Shorten(baseURL string, originalURL string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("url is empty: %w ", ErrInvalidURL)
	}

	const maxAttempts = 10 // Максимальное количество попыток
	attempts := 0

	for attempts < maxAttempts {
		id, err := s.repo.Save(originalURL)

		if errorsUtils.IsUniqueViolation(err) {
			return baseURL + "/" + id, fmt.Errorf("url %s is not unique: %w ", originalURL, err)
		}

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

// func (s *URLService) BeginTransaction() (repository.Transaction, error) {
// 	var transactionRepo repository.TransactableRepository

// 	if postgresRepo, ok := s.repo.(repository.TransactableRepository); ok {
// 		transactionRepo = postgresRepo
// 	} else {
// 		return nil, fmt.Errorf("transactionRepo  does not implement ITransactableRepository: %w ", ErrInvalidURL)
// 	}

// 	tx, err := transactionRepo.Begin()
// 	if err != nil {
// 		return nil, fmt.Errorf("tx begin error: %w ", ErrInvalidURL)
// 	}

// 	return tx, nil
// }
