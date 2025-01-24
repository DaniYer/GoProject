package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type URLShortener struct {
	store map[string]string
	mu    sync.RWMutex
	rng   *rand.Rand
}

func NewURLShortener() *URLShortener {
	// Создаём новый локальный генератор случайных чисел
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	return &URLShortener{
		store: make(map[string]string),
		rng:   rng,
	}
}

func (s *URLShortener) generateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const idLength = 8

	var sb strings.Builder
	for i := 0; i < idLength; i++ {
		sb.WriteByte(charset[s.rng.Intn(len(charset))])
	}
	return sb.String()
}

func (s *URLShortener) shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, "Empty URL", http.StatusBadRequest)
		return
	}

	id := s.generateID()
	s.mu.Lock()
	s.store[id] = originalURL
	s.mu.Unlock()

	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", id)
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortenedURL))
}

func (s *URLShortener) resolveURLHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	originalURL, exists := s.store[id]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func main() {
	shortener := NewURLShortener()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			shortener.shortenURLHandler(w, r)
		} else {
			shortener.resolveURLHandler(w, r)
		}
	})

	fmt.Println("Server is running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
}
