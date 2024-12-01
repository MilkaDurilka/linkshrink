package controller

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"linkshrink/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
		"github.com/gorilla/mux"
)

// MockURLService - мок-сервис для тестирования с использованием testify
type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) Shorten(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockURLService) GetOriginalURL(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func TestShortenURL(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockShorten    func(m *MockURLService)
		expectedCode   int
		expectedBody   string
	}{
		{
			name:         "Valid URL",
			body:         "http://example.com",
			mockShorten: func(m *MockURLService) {
				m.On("Shorten", "http://example.com").Return("short.ly/abc123", nil)
			},
			expectedCode: http.StatusCreated,
			expectedBody: "short.ly/abc123",
		},
		{
			name:         "Empty URL",
			body:         "",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid URL\n",
		},
		{
			name: "Invalid URL",
			body: "http://invalid-url",
			mockShorten: func(m *MockURLService) {
				m.On("Shorten", "http://invalid-url").Return("", service.ErrInvalidURL)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid URL\n",
		},
		{
			name: "Internal Server Error",
			body: "http://example.com",
			mockShorten: func(m *MockURLService) {
				m.On("Shorten", "http://example.com").Return("", errors.New("some error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Error shortening URL\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockURLService)
			if tt.mockShorten != nil {
				tt.mockShorten(mockService)
			}
			controller := NewURLController(mockService)

			req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			controller.ShortenURL(rr, req)

			res := rr.Result()
			assert.Equal(t, tt.expectedCode, res.StatusCode)

			body, _ := io.ReadAll(res.Body)
			err := res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.expectedBody, string(body))

			mockService.AssertExpectations(t)
		})
	}
}

func TestRedirectURL(t *testing.T) {
	tests := []struct {
		name             string
		id               string
		mockGetOriginal  func(m *MockURLService)
		expectedCode     int
		expectedLocation string
	}{
		{
			name: "Valid ID",
			id:   "abc123",
			mockGetOriginal: func(m *MockURLService) {
				m.On("GetOriginalURL", "abc123").Return("http://example.com", nil)
			},
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "http://example.com",
		},
		{
			name: "URL Not Found",
			id:   "nonexistent",
			mockGetOriginal: func(m *MockURLService) {
				m.On("GetOriginalURL", "nonexistent").Return("", service.ErrURLNotFound)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Internal Server Error",
			id:   "abc123",
			mockGetOriginal: func(m *MockURLService) {
				m.On("GetOriginalURL", "abc123").Return("", errors.New("some error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockURLService)
			if tt.mockGetOriginal != nil {
				tt.mockGetOriginal(mockService)
			}
			controller := NewURLController(mockService)
			r := mux.NewRouter()
	    r.HandleFunc("/{id}", controller.RedirectURL)


			req := httptest.NewRequest("GET", "/"+tt.id, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			res := rr.Result()
			
			err := res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCode, res.StatusCode)

			if tt.expectedCode == http.StatusTemporaryRedirect {
				location := res.Header.Get("Location")
				assert.Equal(t, tt.expectedLocation, location)
			}

			mockService.AssertExpectations(t)
		})
	}
}
