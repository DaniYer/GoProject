package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateRandomID генерирует случайный строковый идентификатор длиной 7 символов.
// Используется криптографически безопасный генератор случайных чисел (crypto/rand),
// что гарантирует высокую степень случайности.
//
// Алгоритм:
//  1. Создаётся срез случайных байт длиной length (7).
//  2. Байты кодируются в шестнадцатеричную строку (hex).
//  3. Обрезается строка до 7 символов (т.к. hex кодирование даёт больше символов, чем байт).
//
// Возвращает:
//   - Строку длиной 7 символов, содержащую 0-9 и a-f.
//   - Пустую строку, если произошла ошибка при генерации случайных байт.
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
