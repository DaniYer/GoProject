package shortener

import (
	"io"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	generaterandomid "github.com/DaniYer/GoProject.git/internal/app/randomid"
	"go.uber.org/zap"
)

func NewGenerateShortURLHandler(cfg *config.Config, store URLStoreWithDBforHandler, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		GenerateShortURLHandler(w, r, cfg, store, logger)
	}
}

func GenerateShortURLHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, store URLStoreWithDBforHandler, logger *zap.SugaredLogger) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)

	existingShortURL, err := store.GetByOriginalURL(originalURL)
	if err == nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(cfg.BaseURL + "/" + existingShortURL))
		return
	}

	shortID := generaterandomid.GenerateRandomID()

	if err := store.Save(shortID, originalURL); err != nil {
		logger.Errorf("Ошибка сохранения: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(cfg.BaseURL + "/" + shortID))
}
