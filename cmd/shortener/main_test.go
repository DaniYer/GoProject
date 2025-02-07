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
				body:   "localhost:8080/1111111", // Пример правильного ответа, который возвращает сервер
				method: "POST",
			}},
		{name: "Empty Body",
			want: want{
				code:   400,
				body:   "",
				method: "POST",
			}},
		{name: "Incorrect Method",
			want: want{
				code:   400,
				body:   "",
				method: "GET",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.want.method, "/", strings.NewReader(tt.want.body))
			w := httptest.NewRecorder()
			writeDate(w, r)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			// Если нужно, добавьте проверку тела ответа
			if tt.want.code == 201 {
				expectedBody := "localhost:8080/1111111"
				actualBody := w.Body.String()
				assert.Equal(t, expectedBody, actualBody)
			}
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
			id:   "1111111",
			want: want{
				code:   http.StatusMovedPermanently,
				method: "GET",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Добавьте запись в storage перед тестом
			storage[tt.id] = "https://example.com"
			r := httptest.NewRequest(tt.want.method, "/"+tt.id, nil)
			w := httptest.NewRecorder()
			redirectedHandler(w, r)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			// Если нужно, добавьте проверку редиректа
			expectedLocation := "https://example.com"
			location := res.Header.Get("Location")
			assert.Equal(t, expectedLocation, location)
		})
	}
}
