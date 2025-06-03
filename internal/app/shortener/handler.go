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

// структура для JSON-запроса
type shortenRequest struct {
	URL string `json:"url"`
}

// структура для JSON-ответа
type shortenResponse struct {
	Result string `json:"result"`
}

func NewHandleShortenURL(cfg *config.Config, write URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleShortenURL(w, r, cfg, write)
	}
}

// HandleShortenURL обрабатывает POST-запрос на сокращение URL
func HandleShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, write URLStore) {

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
	if err := write.Save(shortURL, req.URL); err != nil {
		sugar.Errorf("Ошибка сохранения в хранилище: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}
	resp := redirectURL{
		Result: "http://localhost:8080/" + shortURL,
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
