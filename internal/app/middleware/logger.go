package middleware

import (
	"net/http"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK, // по умолчанию 200
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logger.Log.Infof(
			"%s %s %d %dB %v",
			r.Method,
			r.RequestURI,
			rw.status,
			rw.size,
			duration,
		)
	})
}
