package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/utils"
)

func AuthMiddleware(next http.HandlerFunc, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		_, err := r.Cookie(utils.CookieName)

		if err != nil {
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

func generateNewUserID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
