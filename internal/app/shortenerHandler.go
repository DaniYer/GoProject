package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Links struct {
	storage map[string]string
}

// Обработчик для сокращения URL (POST /)
func (a *Links) Shortener(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Чтение URL из тела запроса
	var originalURL string
	_, err := fmt.Fscanf(r.Body, "%s", &originalURL)
	if err != nil || originalURL == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Генерация уникального ID для сокращенного URL
	id := generateID()

	// Сохранение соответствия в мапе
	a.storage[id] = originalURL

	// Создание сокращенной ссылки
	smallURL := fmt.Sprintf("http://localhost:8080/%s", id)

	// Отправка ответа с сокращенным URL
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, smallURL)
}

// Генерация уникального ID для сокращенного URL
func generateID() string {
	return uuid.New().String()[:7] // Генерация ID длиной 7 символов
}

// Обработчик для перенаправления по сокращенному URL (GET /{id})
func (a *Links) ShortenerLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Извлечение ID из URL
	id := getIDFromURL(r.URL.Path)

	// Проверка наличия URL в хранилище
	originalURL, exists := a.storage[id]
	if !exists {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	// Перенаправление на оригинальный URL
	http.Redirect(w, r, originalURL, http.StatusFound) // Код 307 (Temporary Redirect)
}

// Функция для извлечения ID из URL
func getIDFromURL(path string) string {
	// Извлекаем часть пути, которая идет после первого слэша
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}
