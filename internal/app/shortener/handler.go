package shortener

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	generaterandomid "github.com/DaniYer/GoProject.git/internal/app/randomid"
	"go.uber.org/zap"
)

type URLStoreWithDB interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
	SaveWithConflict(shortURL, originalURL string) (string, error)
}

type URLStoreWithDBforHandler interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
	SaveWithConflict(shortURL, originalURL string) (string, error)
	GetByOriginalURL(originalURL string) (string, error)
}

// Структуры для JSON-запроса и ответа
type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

// Хендлер для /api/shorten (JSON)
func NewHandleShortenURLv13(cfg *config.Config, store URLStoreWithDBforHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleShortenURLv13(w, r, cfg, store, sugar)
	}
}

func HandleShortenURLv13(w http.ResponseWriter, r *http.Request, cfg *config.Config, store URLStoreWithDBforHandler, logger *zap.SugaredLogger) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req shortenRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	existingShortURL, err := store.GetByOriginalURL(req.URL)
	if err == nil {
		resp := shortenResponse{
			Result: cfg.BaseURL + "/" + existingShortURL,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(resp)
		return
	}

	shortID := generaterandomid.GenerateRandomID()

	if err := store.Save(shortID, req.URL); err != nil {
		logger.Errorf("Ошибка сохранения: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resp := shortenResponse{
		Result: cfg.BaseURL + "/" + shortID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
