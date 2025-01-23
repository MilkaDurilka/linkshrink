package service

import (
	"context"
	"errors"
	"fmt"
	"linkshrink/internal/repository"
	"linkshrink/internal/utils"
	errorsUtils "linkshrink/internal/utils/errors"
	"linkshrink/internal/utils/transaction"
)

var (
	ErrInvalidURL     = errors.New("invalid URL")
	ErrURLNotFound    = errors.New("URL not found")
	ErrInternalServer = errors.New("internal Server Error")
)

type IURLService interface {
	Shorten(ctx context.Context, baseURL string, url string) (string, error)
	GetOriginalURL(id string) (string, error)
	BatchShorten(
		ctx context.Context,
		baseURL string,
		params []utils.BatchShortenParam,
	) ([]utils.BatchShortenReturnParam, error)
}

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{ // Возвращаем новый сервис с заданным репозиторием
		repo: repo,
	}
}

// BatchShorten сокращает оригинальный URL.
func (s *URLService) BatchShorten(
	ctx context.Context,
	baseURL string,
	params []utils.BatchShortenParam,
) ([]utils.BatchShortenReturnParam, error) {
	txManager, ok := ctx.Value(utils.TxManager).(transaction.TxManager)
	var response []utils.BatchShortenReturnParam
	if ok {
		err := txManager.ReadCommitted(ctx, func(ctx context.Context) error {
			var errTx error
			response, errTx = s.batchShortenInner(ctx, baseURL, params)
			return errTx
		})

		if err != nil {
			return nil, fmt.Errorf("txManager ReadCommitted batchShortenInner: %w ", err)
		}

		return response, nil
	}

	return s.batchShortenInner(ctx, baseURL, params)
}

func (s *URLService) batchShortenInner(
	ctx context.Context,
	baseURL string,
	params []utils.BatchShortenParam,
) ([]utils.BatchShortenReturnParam, error) {
	res := make([]utils.BatchShortenReturnParam, 0, len(params))
	for _, param := range params {
		if len(param.OriginalURL) == 0 {
			return nil, fmt.Errorf("url %s is empty: %w ", param.CorrelationID, ErrInvalidURL)
		}
		url, err := s.Shorten(ctx, baseURL, param.OriginalURL)

		if err != nil {
			return nil, fmt.Errorf("batchShortenInner: %w ", err)
		}

		res = append(res, utils.BatchShortenReturnParam{
			CorrelationID: param.CorrelationID,
			ShortURL:      url,
		})
	}

	return res, nil
}

// Shorten сокращает оригинальный URL.
func (s *URLService) Shorten(ctx context.Context, baseURL string, originalURL string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("url is empty: %w ", ErrInvalidURL)
	}

	const maxAttempts = 10 // Максимальное количество попыток
	attempts := 0

	for attempts < maxAttempts {
		id, err := s.repo.Save(ctx, originalURL)

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
