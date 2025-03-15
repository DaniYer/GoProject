package gzipmiddleware

import (
	"bytes"
	"compress/gzip"
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

	})
}
