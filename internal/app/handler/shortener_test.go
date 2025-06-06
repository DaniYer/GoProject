package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/repository"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/stretchr/testify/assert"
)

// Создаем полностью in-memory хранилище для тестов
func createTestHandler() *Handler {
	repo := repository.NewInMemoryRepository()
	svc := service.NewService(repo)

	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	// в тестах передаем заглушку pingFunc
	return NewHandler(svc, cfg, func(ctx context.Context) error { return nil })
}

func TestShortenHandler(t *testing.T) {
	h := createTestHandler()

	body := bytes.NewBufferString("https://practicum.yandex.ru/")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	w := httptest.NewRecorder()

	h.ShortenHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
}

func TestRedirectHandler(t *testing.T) {
	h := createTestHandler()

	// Сохраняем заранее запись, чтобы протестировать редирект
	userID := "testuser"
	originalURL := "https://practicum.yandex.ru/"
	shortID, _ := h.service.SaveURL(context.Background(), userID, originalURL)

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	w := httptest.NewRecorder()

	h.RedirectHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, originalURL, res.Header.Get("Location"))
}
