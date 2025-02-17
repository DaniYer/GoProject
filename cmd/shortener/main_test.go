package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_shortenedURL(t *testing.T) {
	type want struct {
		code   int
		method string
	}
	tests := []struct {
		name   string
		body   string
		method string
		want   want
	}{
		{
			name:   "Positive Test - Valid URL",
			body:   "https://practicum.yandex.ru/",
			method: "POST",
			want: want{
				code:   http.StatusCreated,
				method: "POST",
			},
		},
		{
			name:   "Negative Test - Empty Body",
			body:   "",
			method: "POST",
			want: want{
				code:   http.StatusBadRequest,
				method: "POST",
			},
		},
		{
			name:   "Negative Test - Incorrect Method",
			body:   "https://practicum.yandex.ru/",
			method: "GET",
			want: want{
				code:   http.StatusBadRequest,
				method: "GET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			shortenedURL(w, r, "http://localhost:8080", "test_storage.json")
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			res.Body.Close()
		})
	}
}

func Test_redirectedURL(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name string
		id   string
		want want
	}{
		{
			name: "Positive Test - Valid Short URL",
			id:   "1111111",
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		},
		{
			name: "Negative Test - Invalid Short URL",
			id:   "unknown",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Предзаполняем хранилище для положительного теста
			storage[tt.id] = "https://example.com"

			r := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
			w := httptest.NewRecorder()

			redirectedURL(w, r)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			res.Body.Close()
		})
	}
}

func Test_apiShortenHandler(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		body   string
		method string
		want   want
	}{
		{
			name:   "Positive Test - Valid JSON",
			body:   `{"url":"https://practicum.yandex.ru/"}`,
			method: "POST",
			want: want{
				code: http.StatusCreated,
			},
		},
		{
			name:   "Negative Test - Invalid JSON",
			body:   `{"url":}`,
			method: "POST",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "Negative Test - Empty Body",
			body:   "",
			method: "POST",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			apiShortenHandler(w, r, "http://localhost:8080", "test_storage.json")

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			if res.StatusCode == http.StatusCreated {
				var resp struct {
					Result string `json:"result"`
				}
				err := json.NewDecoder(res.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Contains(t, resp.Result, "http://localhost:8080/")
			}

			res.Body.Close()
		})
	}
}
