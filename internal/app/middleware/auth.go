package middleware

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/utils"
)

func AuthMiddleware(next http.HandlerFunc, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		if _, err := r.Cookie(utils.CookieName); err != nil {
			userID = generateNewUserID()
			newCookie := utils.GenerateSignedCookie(userID, secret)
			http.SetCookie(w, newCookie)
		} else {
			userID, err = utils.ValidateCookie(r, secret)
			if err != nil {
				userID = generateNewUserID()
				newCookie := utils.GenerateSignedCookie(userID, secret)
				http.SetCookie(w, newCookie)
			}
		}

		ctx := utils.WithUserID(r.Context(), userID)
		next(w, r.WithContext(ctx))
	})
}

// generateNewUserID создает новый уникальный идентификатор пользователя
// Использует текущее время в наносекундах и случайное число для обеспечения уникальности
// Это простой способ, но в реальных приложениях лучше использовать UUID или другие методы генерации уникальных идентификаторов
// Для тестов можно использовать mock-функцию
func generateNewUserID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.Itoa(rand.Intn(1000))
}
