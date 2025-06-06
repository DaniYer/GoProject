package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipDecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Bad Gzip Content", http.StatusBadRequest)
				return
			}
			defer gr.Close()

			r.Body = io.NopCloser(gr)
		}
		next.ServeHTTP(w, r)
	})
}
