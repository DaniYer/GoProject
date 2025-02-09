package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/go-chi/chi/v5"
)

var storage = map[string]string{}

func main() {
	// Получаем конфигурацию из пакета config
	cfg := config.NewConfig()

	// Выводим конфигурацию (для отладки)
	cfg.Print()
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		shortenedURL(w, r, cfg)
	})
	r.Get("/{id}", redirectedURL)
	fmt.Println(cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		panic(err)
	}
}
func shortenedURL(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Только POST запросы
	if r.Method != http.MethodPost {
		http.Error(w, "Unresolved method", 400)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	url := string(body)

	// Генерация уникального идентификатора
	genID := genSym()

	// Сохранение в мапу
	storage[genID] = url

	// Отправка ответа с сокращенным URL
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%s/%s", cfg.BaseURL, genID)))
}

func redirectedURL(w http.ResponseWriter, r *http.Request) {
	// Только GET запросы
	if r.Method != http.MethodGet {
		http.Error(w, "Unresolved method", 400)
		return
	}
	// Извлечение идентификатора из пути
	id := r.URL.Path[1:]
	if id == "" {
		http.Error(w, "Empty path or not found", 400)
		return
	}
	// Поиск в мапе
	value, exists := storage[id]
	if !exists {
		http.Error(w, "URL not found", 400)
		return
	}
	// Редирект на оригинальный URL
	http.Redirect(w, r, value, http.StatusTemporaryRedirect)
	w.Header().Set("Content-Type", "text/plain")
}
func genSym() string {
	result := ""
	for i := 0; i < 8; i++ {
		result += string(rune(rand.Intn(26) + 'a')) // Генерация случайных символов
	}
	return result
}
