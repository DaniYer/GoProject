package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

// contextKey — собственный тип для ключей контекста, чтобы избежать коллизий.
type contextKey string

// UserIDKey — ключ для хранения userID в контексте запроса.
const UserIDKey contextKey = "userID"

// secretKey — секрет для подписи userID в cookie.
// Для учебного проекта можно захардкодить, в проде лучше хранить в переменных окружения.
const secretKey = "practicum_secret_key"

// AuthMiddleware — промежуточный обработчик, который:
// 1. Проверяет, есть ли кука с userID.
// 2. Если нет — генерирует новый userID и ставит куку.
// 3. Если есть, проверяет подпись.
// 4. Если подпись невалидна — генерирует новый userID.
// 5. Кладёт userID в контекст, чтобы его могли использовать обработчики дальше.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Пытаемся достать куку "auth"
		cookie, err := r.Cookie("auth")
		if err != nil {
			// Кука не найдена → создаём нового пользователя
			userID := generateUserID()
			signed := sign(userID)

			// Сохраняем userID + подпись в куку
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

		// Разбиваем куку на userID и подпись
		parts := strings.Split(cookie.Value, "|")
		if len(parts) != 2 {
			// Формат неверный → генерируем нового пользователя
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

		// Проверяем подпись
		if sign(userID) != signature {
			// Подпись неверная → генерируем нового пользователя
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

		// Всё ок — userID достали и проверили
		log.Printf("[AUTH] Распарсили userID из куки: %s", userID)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// sign — создаёт SHA256-хеш userID + secretKey, возвращает hex-строку.
// Используется для защиты куки от подделки.
func sign(userID string) string {
	hash := sha256.Sum256([]byte(userID + secretKey))
	return hex.EncodeToString(hash[:])
}

// generateUserID — генерирует случайный 16-символьный идентификатор
// из букв и цифр.
func generateUserID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[rand.Intn(36)]
	}
	return string(b)
}
