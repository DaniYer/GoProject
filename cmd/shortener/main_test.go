package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_writeDate(t *testing.T) {

	type want struct {
		code   int
		body   string
		method string
	}

	tests := []struct {
		name string
		want want
	}{
		{name: "positive test",
			want: want{
				code:   201,
				body:   "https://practicum.yandex.ru/", // Пример тела для POST запроса
				method: "POST",
			}},
		{name: "Empty Body",
			want: want{
				code:   400,
				body:   "", // Пустое тело запроса
				method: "POST",
			}},
		{name: "Incorrect Method",
			want: want{
				code:   400,
				body:   "",    // Пустое тело запроса
				method: "GET", // Неверный метод
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.want.method, "localhost:8080", strings.NewReader(tt.want.body))
			w := httptest.NewRecorder()
			writeDate(w, r)
			res := w.Result()
			res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)

		})
	}
}

func Test_redirectedHandler(t *testing.T) {

	type want struct {
		code   int
		method string
	}
	tests := []struct {
		name string
		id   string
		want want
	}{
		{
			name: "Pos Request",
			id:   "1111111", // ID для редиректа

			want: want{
				code:   http.StatusMovedPermanently, // Ожидаемый статус редиректа
				method: "GET",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage[tt.id] = "example.com" // Добавление URL в storage перед тестом
			r := httptest.NewRequest(tt.want.method, "/"+tt.id, nil)
			w := httptest.NewRecorder()
			redirectedHandler(w, r)
			res := w.Result()
			res.Body.Close()

			// Здесь также не нужно закрывать res.Body
			// defer r.Body.Close() — это важно, так как респонс не требует закрытия
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
