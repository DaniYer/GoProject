package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenURLHandler(t *testing.T) {
	shortener := NewURLShortener()

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedPrefix string
	}{
		{
			name:           "Valid URL",
			method:         http.MethodPost,
			body:           "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedPrefix: "http://localhost:8080/",
		},
		{
			name:           "Empty Body",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			shortener.shortenURLHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedPrefix != "" {
				body, _ := io.ReadAll(resp.Body)
				if !strings.HasPrefix(string(body), tt.expectedPrefix) {
					t.Errorf("expected response to start with %s, got %s", tt.expectedPrefix, string(body))
				}
			}
		})
	}
}

func TestResolveURLHandler(t *testing.T) {
	shortener := NewURLShortener()
	originalURL := "https://example.com"
	id := shortener.generateID()

	// Добавляем URL в хранилище
	shortener.mu.Lock()
	shortener.store[id] = originalURL
	shortener.mu.Unlock()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Valid ID",
			path:           "/" + id,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    originalURL,
		},
		{
			name:           "Invalid ID",
			path:           "/invalidID",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty ID",
			path:           "/",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			shortener.resolveURLHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == http.StatusTemporaryRedirect {
				location := resp.Header.Get("Location")
				if location != tt.expectedURL {
					t.Errorf("expected redirect to %s, got %s", tt.expectedURL, location)
				}
			}
		})
	}