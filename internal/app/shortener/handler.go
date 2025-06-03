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

type URLStoreWithDBforHandler interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
	SaveWithConflict(shortURL, originalURL string) (string, error)
	GetByOriginalURL(originalURL string) (string, error)
}

// структура для JSON-запроса
type shortenRequest struct {
	URL string `json:"url"`
}

// структура для JSON-ответа
type shortenResponse struct {
	Result string `json:"result"`
}

func NewHandleShortenURLv13(cfg *config.Config, write URLStoreWithDBforHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleShortenURLv13(w, r, cfg, write, sugar)
	}
}

// HandleShortenURL обрабатывает POST-запрос на сокращение URL
func HandleShortenURLv13(w http.ResponseWriter, r *http.Request, cfg *config.Config, store URLStoreWithDBforHandler, logger *zap.SugaredLogger) {
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

	// Пытаемся проверить, есть ли уже такой URL в БД:
	existingShortURL, err := store.GetByOriginalURL(req.URL)
	if err == nil {
		// Нашли дубликат — возвращаем 409
		resp := redirectURL{
			Result: cfg.BaseURL + "/" + existingShortURL,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Создаем новый короткий URL
	shortURL := generaterandomid.GenerateRandomID()

	if err := store.Save(shortURL, req.URL); err != nil {
		logger.Errorf("Ошибка сохранения: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resp := redirectURL{
		Result: cfg.BaseURL + "/" + shortURL,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

type shortenURL struct {
	URL string `json:"url"`
}

type redirectURL struct {
	Result string `json:"result"`
}

func NewHandleShortenURLv7(cfg *config.Config, write URLStoreWithDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleShortenURLv7(w, r, cfg, write, sugar)
	}
}

func HandleShortenURLv7(w http.ResponseWriter, r *http.Request, cfg *config.Config, store URLStoreWithDB, logger *zap.SugaredLogger) {
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

	// Просто создаём новый короткий URL без проверок дубликатов
	shortURL := generaterandomid.GenerateRandomID()

	if err := store.Save(shortURL, req.URL); err != nil {
		logger.Errorf("Ошибка сохранения: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resp := shortenResponse{
		Result: cfg.BaseURL + "/" + shortURL,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
