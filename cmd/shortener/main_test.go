package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortenURLHandler(t *testing.T) {
	shortener := NewURLShortener()

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedResult string
	}{
		{
			name:           "Valid URL",
			method:         http.MethodPost,
			body:           "https://example.com",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Empty body",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			shortener.shortenURLHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusCreated {
				assert.Contains(t, rec.Body.String(), "http://localhost:8080/")
			}
		})
	}
}

func TestResolveURLHandler(t *testing.T) {
	shortener := NewURLShortener()

	// Добавляем тестовые данные
	shortener.mu.Lock()
	shortener.store["testID"] = "https://example.com"
	shortener.mu.Unlock()

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Valid ID",
			id:             "testID",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    "https://example.com",
		},
		{
			name:           "Invalid ID",
			id:             "invalidID",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
			rec := httptest.NewRecorder()

			shortener.resolveURLHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.expectedURL, rec.Header().Get("Location"))
			}
		})
	}
}
