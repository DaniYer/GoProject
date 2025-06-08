package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateRandomID() string {
	length := 7
	// генерируем случайный слайс байт
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return ""
	}
	return hex.EncodeToString(randomBytes)[:length]
}
