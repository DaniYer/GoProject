// package main

// import (
// 	"fmt"
// 	"io"
// 	"math/rand"
// 	"net/http"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/mux"
// )

// type URLShortener struct {
// 	store map[string]string
// 	mu    sync.RWMutex
// 	rng   *rand.Rand
// }

// func NewURLShortener() *URLShortener {
// 	source := rand.NewSource(time.Now().UnixNano())
// 	rng := rand.New(source)

// 	return &URLShortener{
// 		store: make(map[string]string),
// 		rng:   rng,
// 	}
// }

// // Генерация случайного идентификатора
// func (s *URLShortener) generateID() string {
// 	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
// 	const idLength = 8

// 	var sb strings.Builder
// 	for i := 0; i < idLength; i++ {
// 		sb.WriteByte(charset[s.rng.Intn(len(charset))])
// 	}
// 	return sb.String()
// }

// // Обработчик для POST-запросов
// func (s *URLShortener) shortenURLHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil || len(body) == 0 {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()

// 	originalURL := strings.TrimSpace(string(body))
// 	if originalURL == "" {
// 		http.Error(w, "Empty URL", http.StatusBadRequest)
// 		return
// 	}

// 	id := s.generateID()
// 	s.mu.Lock()
// 	s.store[id] = originalURL
// 	s.mu.Unlock()

// 	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", id)
// 	w.WriteHeader(http.StatusCreated)
// 	_, _ = w.Write([]byte(shortenedURL))
// }

// // Обработчик для GET-запросов
// func (s *URLShortener) resolveURLHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	id := vars["id"]

// 	if id == "" {
// 		http.Error(w, "ID not provided", http.StatusBadRequest)
// 		return
// 	}

// 	s.mu.RLock()
// 	originalURL, exists := s.store[id]
// 	s.mu.RUnlock()

// 	if !exists {
// 		http.Error(w, "URL not found", http.StatusNotFound)
// 		return
// 	}

// 	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
// }

// func main() {
// 	shortener := NewURLShortener()
// 	r := mux.NewRouter()

// 	// Настройка маршрутов
// 	r.HandleFunc("/", shortener.shortenURLHandler).Methods(http.MethodPost)
// 	r.HandleFunc("/{id}", shortener.resolveURLHandler).Methods(http.MethodGet)

// 	fmt.Println("Server is running at http://localhost:8080")
// 	if err := http.ListenAndServe(":8080", r); err != nil {
// 		fmt.Println("Error starting server:", err)
// 	}
// }

package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

var storage = map[string]string{}

func main() {
	http.HandleFunc("/", shortenedURL)      // Обработка POST запроса на сокращение URL
	http.HandleFunc("/{id}", redirectedURL) // Обработка GET запроса на редирект
	fmt.Println("localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func shortenedURL(w http.ResponseWriter, r *http.Request) {
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
	genId := genSym()

	// Сохранение в мапу
	storage[genId] = url

	// Отправка ответа с сокращенным URL
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("http://localhost:8080/" + genId))
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
