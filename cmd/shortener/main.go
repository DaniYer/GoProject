package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type URLShortener struct {
	store map[string]string
	mu    sync.RWMutex
	rng   *rand.Rand
}

func NewURLShortener() *URLShortener {
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

// Handler for shortening URL
func (s *URLShortener) shortenURLHandler(w http.ResponseWriter, r *http.Request) {
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

// Handler for resolving URL
func (s *URLShortener) resolveURLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

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

	// Create a new router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/", shortener.shortenURLHandler).Methods(http.MethodPost)
	r.HandleFunc("/{id}", shortener.resolveURLHandler).Methods(http.MethodGet)

	// Start the server
	fmt.Println("Server is running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
