package shortener

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	generaterandomid "github.com/DaniYer/GoProject.git/internal/app/randomid"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

type URLStoreWithDB interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
	SaveWithConflict(shortURL, originalURL string) (string, error)
}

// структура для JSON-запроса
type shortenRequest struct {
	URL string `json:"url"`
}

// структура для JSON-ответа
type shortenResponse struct {
	Result string `json:"result"`
}

func NewHandleShortenURL(cfg *config.Config, write URLStoreWithDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleShortenURL(w, r, cfg, write)
	}
}

// HandleShortenURL обрабатывает POST-запрос на сокращение URL
func HandleShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, write URLStoreWithDB) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req shortenURL
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	shortURL := generaterandomid.GenerateRandomID()

	// Пытаемся сохранить
	existingShortURL, err := write.SaveWithConflict(shortURL, req.URL)
	if err != nil {
		sugar.Errorf("Ошибка сохранения в хранилище: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resp := redirectURL{
		Result: cfg.BaseURL + "/" + existingShortURL,
	}

	w.Header().Set("Content-Type", "application/json")

	if existingShortURL != shortURL {
		// Уже существующий URL
		w.WriteHeader(http.StatusConflict)
	} else {
		// Новый URL
		w.WriteHeader(http.StatusCreated)
	}

	json.NewEncoder(w).Encode(resp)
}

type shortenURL struct {
	URL string `json:"url"`
}

type redirectURL struct {
	Result string `json:"result"`
}
