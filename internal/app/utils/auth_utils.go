package utils

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

const (
	CookieName = "auth"
)

type ctxKey string

var userIDKey = ctxKey("userID")

func GenerateSignedCookie(userID string, secret string) *http.Cookie {
	sign := generateHMAC(userID, secret)
	value := userID + ":" + sign

	return &http.Cookie{
		Name:  CookieName,
		Value: value,
		Path:  "/",
	}
}

func ValidateCookie(r *http.Request, secret string) (string, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(cookie.Value, ":", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid cookie format")
	}

	userID := parts[0]
	sign := parts[1]

	if sign != generateHMAC(userID, secret) {
		return "", errors.New("invalid cookie signature")
	}

	return userID, nil
}

func generateHMAC(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) string {
	val := ctx.Value(userIDKey)
	if val == nil {
		return ""
	}
	return val.(string)
}
