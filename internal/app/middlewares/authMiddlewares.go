package middlewares

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey  contextKey = "userID"
	CookieName string     = "auth"
)

var secretKey = []byte("super_secret_key") // вынеси потом в config/env

func generateUserID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func sign(value string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

func verify(value, signature string) bool {
	return hmac.Equal([]byte(sign(value)), []byte(signature))
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil || !validateCookie(cookie) {
			userID := generateUserID()
			sig := sign(userID)

			http.SetCookie(w, &http.Cookie{
				Name:  CookieName,
				Value: userID + "|" + sig,
				Path:  "/",
			})

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userID := parseCookie(cookie)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateCookie(cookie *http.Cookie) bool {
	parts := strings.Split(cookie.Value, "|")
	if len(parts) != 2 {
		return false
	}
	return verify(parts[0], parts[1])
}

func parseCookie(cookie *http.Cookie) string {
	parts := strings.Split(cookie.Value, "|")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}
