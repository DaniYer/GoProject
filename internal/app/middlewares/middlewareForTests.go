package middlewares

import (
	"context"
	"net/http"
)

func InjectTestUserIDMiddleware(testUserID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), UserIDKey, testUserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
