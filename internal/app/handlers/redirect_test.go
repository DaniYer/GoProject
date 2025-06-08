package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// Мок URLStore для тестов
type MockStore struct {
	GetFunc func(shortURL string) (string, error)
}

func (m *MockStore) Get(shortURL string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(shortURL)
	}
	return "", nil
}

func (m *MockStore) Save(shortURL, originalURL string) (string, error) {
	return shortURL, nil // обновленная сигнатура под новую архитектуру
}

func (m *MockStore) GetByOriginalURL(originalURL string) (string, error) {
	return "", nil // не используется в redirect тестах
}

func TestRedirectToOriginalURL_Success(t *testing.T) {
	mockStore := &MockStore{
		GetFunc: func(shortURL string) (string, error) {
			if shortURL == "abcd1234" {
				return "http://example.com", nil
			}
			return "", errors.New("not found")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/abcd1234", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abcd1234")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	RedirectToOriginalURL(rec, req, mockStore)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	location := res.Header.Get("Location")
	assert.Equal(t, "http://example.com", location)
}

func TestRedirectToOriginalURL_NotFound(t *testing.T) {
	mockStore := &MockStore{
		GetFunc: func(shortURL string) (string, error) {
			return "", errors.New("not found")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "unknown")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	RedirectToOriginalURL(rec, req, mockStore)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRedirectToOriginalURL_EmptyID(t *testing.T) {
	mockStore := &MockStore{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	RedirectToOriginalURL(rec, req, mockStore)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRedirectToOriginalURL_InvalidMethod(t *testing.T) {
	mockStore := &MockStore{}

	req := httptest.NewRequest(http.MethodPost, "/abcd1234", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abcd1234")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	RedirectToOriginalURL(rec, req, mockStore)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
