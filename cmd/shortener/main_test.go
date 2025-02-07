package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_shortenedURL(t *testing.T) {

	type want struct {
		code   int
		body   string
		method string
	}

	tests := []struct {
		name string
		want want
	}{
		// TODO: Add test cases.
		{name: "positive test",
			want: want{
				code:   201,
				body:   "https://practicum.yandex.ru/",
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
			r := httptest.NewRequest(tt.want.method, "localhost:8080", strings.NewReader(tt.want.body))
			w := httptest.NewRecorder()
			shortenedURL(w, r)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			res.Body.Close()

		})
	}
}

func Test_redirectedURL(t *testing.T) {

	type want struct {
		code int

		method string
	}
	tests := []struct {
		name string
		id   string
		want want
	}{
		// TODO: Add test cases.
		{
			name: "Pos Request",
			id:   "1111111",

			want: want{307, "GET"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage[tt.id] = "expample.com"
			r := httptest.NewRequest(tt.want.method, "/"+tt.id, nil)
			w := httptest.NewRecorder()
			redirectedURL(w, r)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			res.Body.Close()

		})
	}
}
