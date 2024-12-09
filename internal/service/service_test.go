package service_test

import (
	"errors"
	"testing"

	"linkshrink/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository - это мок для репозитория URL.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(originalURL string) (string, error) {
	args := m.Called(originalURL)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) Find(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

// TestURLService_Shortcut тестирует метод Shorten.
func TestURLService_Shortcut(t *testing.T) {
	mockRepo := new(MockRepository)
	srv := service.NewURLService(mockRepo)

	originalURL := "http://example.com"
	baseURL := "http://localhost:8080/"
	mockRepo.On("Save", originalURL).Return("id", nil)

	shortenedURL, err := srv.Shorten(baseURL, originalURL)

	require.NoError(t, err)
	assert.Contains(t, shortenedURL, "http://localhost:8080/") // Проверяем, что URL содержит базовый адрес
	mockRepo.AssertExpectations(t)
}

// TestURLService_Shortcut_InvalidURL тестирует метод Shorten с недопустимым URL.
func TestURLService_Shortcut_InvalidURL(t *testing.T) {
	mockRepo := new(MockRepository)
	srv := service.NewURLService(mockRepo)
	baseURL := "http://localhost:8080/"

	shortenedURL, err := srv.Shorten(baseURL, "")
	assert.True(t, errors.Is(err, service.ErrInvalidURL), "expected ErrInvalidURL")
	assert.Empty(t, shortenedURL)
}

// TestURLService_GetOriginalURL тестирует метод GetOriginalURL.
func TestURLService_GetOriginalURL(t *testing.T) {
	mockRepo := new(MockRepository)
	srv := service.NewURLService(mockRepo)

	id := "abc123"
	originalURL := "http://example.com"
	mockRepo.On("Find", id).Return(originalURL, nil)

	result, err := srv.GetOriginalURL(id)

	require.NoError(t, err)
	assert.Equal(t, originalURL, result)
	mockRepo.AssertExpectations(t)
}

// TestURLService_GetOriginalURL_NotFound тестирует метод GetOriginalURL с несуществующим ID.
func TestURLService_GetOriginalURL_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	srv := service.NewURLService(mockRepo)

	id := "nonexistent"
	mockRepo.On("Find", id).Return("", service.ErrURLNotFound)

	result, err := srv.GetOriginalURL(id)

	assert.True(t, errors.Is(err, service.ErrURLNotFound), "expected ErrURLNotFound")
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
