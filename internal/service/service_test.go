package service_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "linkshrink/internal/service"
)

// MockRepository - это мок для репозитория URL
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Save(id, originalURL string) error {
    args := m.Called(id, originalURL)
    return args.Error(0)
}

func (m *MockRepository) Find(id string) (string, error) {
    args := m.Called(id)
    return args.String(0), args.Error(1)
}

// TestURLService_Shortcut тестирует метод Shorten
func TestURLService_Shortcut(t *testing.T) {
    mockRepo := new(MockRepository)
    service := service.NewURLService(mockRepo)

    originalURL := "http://example.com"
		baseURL := "http://localhost:8080/"
    mockRepo.On("Save", mock.Anything, originalURL).Return(nil)

    shortenedURL, err := service.Shorten(baseURL, originalURL)

    require.NoError(t, err)
    assert.Contains(t, shortenedURL, "http://localhost:8080/") // Проверяем, что URL содержит базовый адрес
    mockRepo.AssertExpectations(t)
}

// TestURLService_Shortcut_InvalidURL тестирует метод Shorten с недопустимым URL
func TestURLService_Shortcut_InvalidURL(t *testing.T) {
    mockRepo := new(MockRepository)
    srv := service.NewURLService(mockRepo)
		baseURL := "http://localhost:8080/"

    shortenedURL, err := srv.Shorten(baseURL, "")
    assert.Equal(t, service.ErrInvalidURL, err)
    assert.Empty(t, shortenedURL)
}

// TestURLService_GetOriginalURL тестирует метод GetOriginalURL
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

// TestURLService_GetOriginalURL_NotFound тестирует метод GetOriginalURL с несуществующим ID
func TestURLService_GetOriginalURL_NotFound(t *testing.T) {
    mockRepo := new(MockRepository)
    srv := service.NewURLService(mockRepo)

    id := "nonexistent"
    mockRepo.On("Find", id).Return("", service.ErrURLNotFound)

    result, err := srv.GetOriginalURL(id)

    assert.Equal(t, service.ErrURLNotFound, err)
    assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
}
