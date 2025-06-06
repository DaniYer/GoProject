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
		userID, err := utils.ValidateCookie(r, secret)
		if err != nil {
			// Генерация новой куки, если невалидна
			userID = generateNewUserID()
			cookie := utils.GenerateSignedCookie(userID, secret)
			http.SetCookie(w, cookie)
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
