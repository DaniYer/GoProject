package handler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/repository"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/DaniYer/GoProject.git/internal/app/utils"
)

// Handler - структура обработчика, содержащая сервис, конфигурацию и функцию проверки доступности базы данных
type Handler struct {
	service *service.Service
	config  *config.Config
	dbPing  func(ctx context.Context) error
}

// NewHandler создает новый обработчик с сервисом, конфигурацией и функцией проверки доступности базы данных
// dbPing - функция для проверки доступности базы данных, может быть nil, если не требуется проверка
func NewHandler(service *service.Service, config *config.Config, dbPing func(ctx context.Context) error) *Handler {
	return &Handler{
		service: service,
		config:  config,
		dbPing:  dbPing,
	}
}

// GET /ping
// Проверяет доступность базы данных (если реализовано) и возвращает "pong"
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	// проверяем пинг только если это postgres
	if h.dbPing != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.dbPing(ctx); err != nil {
			http.Error(w, "Database not available", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// POST /
func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r.Context())

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortID, err := h.service.SaveURL(context.Background(), userID, string(body))
	if err != nil {
		if err == repository.ErrDuplicateURL {
			resultURL := h.config.BaseURL + "/" + shortID
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(resultURL))
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	resultURL := h.config.BaseURL + "/" + shortID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

// GET /{id}
func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")

	originalURL, err := h.service.GetOriginalURL(context.Background(), id)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if originalURL == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
