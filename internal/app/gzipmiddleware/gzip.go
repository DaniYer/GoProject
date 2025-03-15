package gzipmiddleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// bufferedResponseWriter буферизует ответ и сохраняет статус.
type bufferedResponseWriter struct {
	http.ResponseWriter
	buf         *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

// newBufferedResponseWriter создаёт новый буферизованный ResponseWriter.
func newBufferedResponseWriter(w http.ResponseWriter) *bufferedResponseWriter {
	return &bufferedResponseWriter{
		ResponseWriter: w,
		buf:            new(bytes.Buffer),
	}
}

// WriteHeader сохраняет статус, не отправляя его сразу.
func (b *bufferedResponseWriter) WriteHeader(statusCode int) {
	if !b.wroteHeader {
		b.statusCode = statusCode
		b.wroteHeader = true
	}
}

// Write буферизует данные.
func (b *bufferedResponseWriter) Write(p []byte) (int, error) {
	return b.buf.Write(p)
}

// flush отправляет буфер с компрессией или без неё в зависимости от условий.
func (b *bufferedResponseWriter) flush(r *http.Request) error {
	// Если заголовок Content-Type не установлен, попробуем определить его по содержимому.
	contentType := b.Header().Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(b.buf.Bytes())
	}

	// Определяем, поддерживает ли клиент gzip и соответствует ли тип контента требованиям.
	shouldCompress := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") &&
		(strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html"))

	// Перед отправкой устанавливаем статус. Если WriteHeader не вызывался, то по умолчанию статус 200.
	if !b.wroteHeader {
		b.statusCode = http.StatusOK
	}
	b.ResponseWriter.WriteHeader(b.statusCode)

	if shouldCompress {
		// Устанавливаем заголовок, удаляем Content-Length
		b.Header().Set("Content-Encoding", "gzip")
		b.Header().Del("Content-Length")
		gz := gzip.NewWriter(b.ResponseWriter)
		defer gz.Close()
		_, err := io.Copy(gz, b.buf)
		return err
	}

	// Если сжатие не требуется – просто отправляем буфер.
	_, err := b.ResponseWriter.Write(b.buf.Bytes())
	return err
}

// GzipHandle — middleware, который:
// 1. Если запрос сжат (Content-Encoding: gzip) — декомпрессирует тело запроса.
// 2. Буферизует ответ и, если клиент поддерживает gzip и Content-Type — application/json или text/html, сжимает ответ.
func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Декомпрессия входящего запроса, если он сжат.
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request", http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			r.Body = gzReader
		}

		// Используем буферизованный ResponseWriter.
		brw := newBufferedResponseWriter(w)
		next.ServeHTTP(brw, r)

		// После обработки запроса отправляем ответ с компрессией, если требуется.
		if err := brw.flush(r); err != nil {
			// Можно логировать ошибку здесь.
		}
	})
}
