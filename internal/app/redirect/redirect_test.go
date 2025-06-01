package redirect

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestConsumer создаёт временный файл с тестовыми данными и возвращает *storage.Consumer.
func createTestConsumer(t *testing.T) *storage.Consumer {
	t.Helper()

	// Создаем временный файл.
	tmpFile, err := os.CreateTemp("", "test_events_*.json")
	require.NoError(t, err)

	// Пример тестовых событий в формате JSON Lines.
	events := []string{
		`{"uuid":"1","short_url":"4rSPg8ap","original_url":"http://yandex.ru"}`,
		`{"uuid":"2","short_url":"edVPg3ks","original_url":"http://ya.ru"}`,
		`{"uuid":"3","short_url":"dG56Hqxm","original_url":"http://practicum.yandex.ru"}`,
	}
	for _, line := range events {
		_, err = tmpFile.WriteString(line + "\n")
		require.NoError(t, err)
	}

	// Закрываем файл после записи, чтобы затем открыть для чтения.
	require.NoError(t, tmpFile.Close())

	consumer, err := storage.NewConsumer(tmpFile.Name())
	require.NoError(t, err)

	// Удаляем временный файл после завершения теста.
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return consumer
}

func TestRedirectToOriginalURL_Found(t *testing.T) {
	consumer := createTestConsumer(t)

	// Создаем HTTP-запрос.
	req := httptest.NewRequest(http.MethodGet, "/4rSPg8ap", nil)

	// Создаем chi.RouteContext и добавляем параметр "id" с нужным значением.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "4rSPg8ap")
	// Вставляем rctx в контекст запроса.
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	RedirectToOriginalURL(rec, req, consumer)

	res := rec.Result()
	defer res.Body.Close()

	// Ожидаем, что функция выполнит редирект (HTTP 307).
	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)

	location, err := res.Location()
	require.NoError(t, err)
	// Для shortURL "4rSPg8ap" ожидаем оригинальный URL "http://yandex.ru"
	assert.Equal(t, "http://yandex.ru", location.String())
}

func TestRedirectToOriginalURL_NotFound(t *testing.T) {
	consumer := createTestConsumer(t)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "unknown")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	RedirectToOriginalURL(rec, req, consumer)

	res := rec.Result()
	defer res.Body.Close()

	// Если событие не найдено, ожидаем статус 400.
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "URL not found")
}
