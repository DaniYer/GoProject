package main

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_shortenedURL(t *testing.T) {
	storage, _ = NewStorage("test_storage.json") // Тестовое хранилище
	defer os.Remove("test_storage.json")         // Удаляем файл после тестов

	tests := []struct {
		name   string
		method string
		body   string
		want   int
	}{
		{"Positive Test", "POST", "https://practicum.yandex.ru/", 201},
		{"Empty Body", "POST", "", 400},
		{"Incorrect Method", "GET", "", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			shortenedURL(w, r, "http://localhost:8080")

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}

func Test_redirectedURL(t *testing.T) {
	storage, _ = NewStorage("test_storage.json") // Тестовое хранилище
	defer os.Remove("test_storage.json")         // Удаляем файл после тестов

	// Добавляем тестовые данные
	storage.SaveURL("test123", "https://example.com")

	tests := []struct {
		name string
		id   string
		want int
	}{
		{"Existing URL", "test123", 307},
		{"Non-Existing URL", "notfound", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/"+tt.id, nil)
			w := httptest.NewRecorder()

			redirectedURL(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}
