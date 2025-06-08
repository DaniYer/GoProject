package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type MockRedirectStore struct {
	GetFunc func(shortURL string) (string, error)
}

func (m *MockRedirectStore) Get(shortURL string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(shortURL)
	}
	return "", nil
}

func (m *MockRedirectStore) Save(shortURL, originalURL, userID string) (string, error) {
	return shortURL, nil
}

func (m *MockRedirectStore) GetByOriginalURL(originalURL string) (string, error) {
	return "", nil
}

func (m *MockRedirectStore) GetAllByUser(userID string) ([]dto.UserURL, error) {
	return nil, nil
}

func TestRedirectToOriginalURL_Success(t *testing.T) {
	mockStore := &MockRedirectStore{
		GetFunc: func(shortURL string) (string, error) {
			if shortURL == "abcd1234" {
				return "http://example.com", nil
			}
			return "", errors.New("not found")
		},
	}

	svc := &service.URLService{
		Store:   mockStore,
		BaseURL: "http://localhost:8080",
	}

	router := chi.NewRouter()
	router.Get("/{id}", NewRedirectToOriginalURL(svc))

	req := httptest.NewRequest(http.MethodGet, "/abcd1234", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	location := res.Header.Get("Location")
	assert.Equal(t, "http://example.com", location)
}
