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

func generateNewUserID() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.Itoa(rand.Intn(1000))
}
