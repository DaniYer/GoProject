package middlewares

import (
	"context"
	"net/http"
)

// InjectTestUserIDMiddleware возвращает middleware,
// который подменяет userID в контексте на заданный testUserID.
// Это используется в тестах, чтобы эмулировать авторизованного пользователя,
// минуя реальную логику аутентификации (AuthMiddleware).
//
// Пример использования в тестах:
//
//	router := chi.NewRouter()
//	router.Use(middlewares.InjectTestUserIDMiddleware("test-user-123"))
//	router.Get("/protected", handler)
//
// Теперь все запросы будут приходить в handler с userID = "test-user-123".
func InjectTestUserIDMiddleware(testUserID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Вкладываем тестовый userID в контекст запроса
			ctx := context.WithValue(r.Context(), UserIDKey, testUserID)

			// Передаём управление следующему обработчику
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
