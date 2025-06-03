package shortener

import (
	"io"
	"net/http"
	"net/url"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	generaterandomid "github.com/DaniYer/GoProject.git/internal/app/randomid"
)

func NewGenerateShortURLHandler(cfg *config.Config, write URLStoreWithDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		GenerateShortURLHandler(w, r, cfg, write)
	}
}

func GenerateShortURLHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, write URLStoreWithDB) {
	shortID := generaterandomid.GenerateRandomID()
	// читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if _, err = url.Parse(string(body)); err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Записываем событие, проверяем ошибку записи
	if err := write.Save(shortID, string(body)); err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(cfg.BaseURL + "/" + shortID))
}
