package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

// InitLogger инициализирует глобальный логгер.
// Вызывается один раз в начале работы приложения.
func InitLogger(logger *zap.SugaredLogger) {
	sugar = logger
}

// WithLogging — middleware для логирования HTTP-запросов.
// Логирует:
//   - URI запроса
//   - HTTP-метод
//   - код ответа
//   - время обработки
//   - размер ответа (в байтах)
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаём объект для хранения данных об ответе
		responseData := &responseData{
			status: 0,
			size:   0,
		}

		// Оборачиваем оригинальный ResponseWriter в наш
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		// Передаём управление следующему обработчику
		h.ServeHTTP(&lw, r)

		// После завершения — логируем данные
		duration := time.Since(start)
		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}

	return http.HandlerFunc(logFn)
}

// responseData хранит данные об ответе для логирования.
type responseData struct {
	status int // код HTTP-статуса
	size   int // размер тела ответа в байтах
}

// loggingResponseWriter — враппер для ResponseWriter,
// который перехватывает код статуса и размер ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write перехватывает запись данных, чтобы посчитать размер ответа.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader перехватывает установку HTTP-статуса.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
