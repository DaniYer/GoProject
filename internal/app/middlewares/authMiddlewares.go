package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Этот секрет можно захардкодить для сдачи проекта
const secretKey = "practicum_secret_key"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		if err != nil {
			userID := generateUserID()
			signed := sign(userID)

			http.SetCookie(w, &http.Cookie{
				Name:  "auth",
				Value: userID + "|" + signed,
				Path:  "/",
			})

			log.Printf("[AUTH] Кука не найдена, сгенерирован новый userID: %s", userID)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		parts := strings.Split(cookie.Value, "|")
		if len(parts) != 2 {
			log.Printf("[AUTH] Некорректная кука: %s", cookie.Value)
			userID := generateUserID()
			signed := sign(userID)

			http.SetCookie(w, &http.Cookie{
				Name:  "auth",
				Value: userID + "|" + signed,
				Path:  "/",
			})

			log.Printf("[AUTH] Сгенерирован новый userID: %s", userID)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userID := parts[0]
		signature := parts[1]

		if sign(userID) != signature {
			log.Printf("[AUTH] Некорректная подпись куки: %s", cookie.Value)
			userID = generateUserID()
			signed := sign(userID)

			http.SetCookie(w, &http.Cookie{
				Name:  "auth",
				Value: userID + "|" + signed,
				Path:  "/",
			})

			log.Printf("[AUTH] Сгенерирован новый userID: %s", userID)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		log.Printf("[AUTH] Распарсили userID из куки: %s", userID)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sign(userID string) string {
	hash := sha256.Sum256([]byte(userID + secretKey))
	return hex.EncodeToString(hash[:])
}

func generateUserID() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 16)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[rand.Intn(36)]
	}
	return string(b)
}
