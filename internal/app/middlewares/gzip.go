package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipWriter — враппер над http.ResponseWriter,
// который перезаписывает метод Write, чтобы сжимать данные.
type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

// Write — переопределяем, чтобы писать в gzip.Writer,
// а не напрямую в ResponseWriter.
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipHandle — middleware, который:
// 1. Декодирует входящие gzip-запросы.
// 2. Сжимает исходящий ответ, если клиент поддерживает gzip.
func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Если входящий запрос в gzip — разжимаем тело
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(gz)
		}

		// Если клиент не поддерживает gzip — просто отдаём как есть
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Настраиваем заголовки ответа
		w.Header().Set("Content-Encoding", "gzip")         // говорим, что тело сжато
		w.Header().Set("Content-Type", "application/json") // 👈 добавлено: отдаём JSON по умолчанию

		// Оборачиваем ResponseWriter в gzip.Writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Передаём управление следующему обработчику
		// с подменённым writer, чтобы он писал в сжатый поток.
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
