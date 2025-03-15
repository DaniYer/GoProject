package shortener

import (
	"io"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	generaterandomid "github.com/DaniYer/GoProject.git/internal/app/randomid"
	"github.com/DaniYer/GoProject.git/internal/app/storage"
	"github.com/google/uuid"
)

func GenerateShortURLHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, write *storage.Producer) {
	shortID := generaterandomid.GenerateRandomID()

	// читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	eventID := uuid.New().String()

	event := storage.Event{
		UUID:        eventID,
		ShortURL:    shortID,
		OriginalURL: string(body),
	}

	// Записываем событие, проверяем ошибку записи
	if err := write.WriteEvent(&event); err != nil {
		http.Error(w, "Failed to write event", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(cfg.B + "/" + shortID))
}
