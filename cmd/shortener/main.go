package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var urlMapping = make(map[string]string)

// Функция для создания уникального сокращённого идентификатора
func shortenUrl(originalUrl string) string {
	hash := md5.Sum([]byte(originalUrl))
	shortened := base64.URLEncoding.EncodeToString(hash[:6])
	return shortened
}

// Эндпоинт для сокращения URL
func createShortUrl(w http.ResponseWriter, r *http.Request) {
	// Чтение тела запроса (оригинальный URL)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	originalUrl := string(body)

	// Генерация сокращённого URL
	shortened := shortenUrl(originalUrl)
	urlMapping[shortened] = originalUrl

	// Формируем ответ
	shortUrl := fmt.Sprintf("http://localhost:8080/%s", shortened)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortUrl))
}

// Эндпоинт для редиректа на оригинальный URL по сокращённому идентификатору
func redirectToOriginalUrl(w http.ResponseWriter, r *http.Request) {
	// Извлекаем сокращённый идентификатор из URL
	shortened := strings.TrimPrefix(r.URL.Path, "/")

	// Проверяем, существует ли сокращённый URL в нашем словаре
	originalUrl, found := urlMapping[shortened]
	if !found {
		http.Error(w, "Not Found", http.StatusBadRequest)
		return
	}

	// Перенаправляем на оригинальный URL
	http.Redirect(w, r, originalUrl, http.StatusFound)
}

func main() {
	http.HandleFunc("/", createShortUrl)                 // Обработка POST запроса для создания сокращённого URL
	http.HandleFunc("/redirect/", redirectToOriginalUrl) // Обработка GET запроса для редиректа

	// Запускаем сервер на порту 8080
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
