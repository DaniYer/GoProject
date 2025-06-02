package shortener

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	generaterandomid "github.com/DaniYer/GoProject.git/internal/app/randomid"
	"github.com/DaniYer/GoProject.git/internal/app/storage"
	"github.com/google/uuid"
)

type Storage interface {
	WriteEvent(*storage.Event, *sql.DB) error
}

// структура для JSON-запроса
type shortenRequest struct {
	URL string `json:"url"`
}

// структура для JSON-ответа
type shortenResponse struct {
	Result string `json:"result"`
}

func NewHandleShortenURL(cfg *config.Config, write Storage, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleShortenURL(w, r, cfg, write, db)
	}
}

// HandleShortenURL обрабатывает POST-запрос на сокращение URL
func HandleShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, write Storage, db *sql.DB) {
	var req shortenRequest

	// Декодируем JSON-запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Генерируем короткий идентификатор
	shortID := generaterandomid.GenerateRandomID()

	eventID := uuid.New().String()

	event := storage.Event{
		UUID:        eventID,
		ShortURL:    shortID,
		OriginalURL: req.URL,
	}
	// Записываем событие, проверяем ошибку записи
	if err := write.WriteEvent(&event, db); err != nil {
		http.Error(w, "Failed to write event", http.StatusInternalServerError)
		return
	}
	// Создаем JSON-ответ
	resp := shortenResponse{
		Result: cfg.B + "/" + shortID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
