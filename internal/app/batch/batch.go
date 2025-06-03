package batch

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/randomid"
)

type URLStore interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewBatchShortenURLHandler(baseURL string, store URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		batchShortenHandler(w, r, baseURL, store)
	}
}
func batchShortenHandler(w http.ResponseWriter, r *http.Request, baseURL string, store URLStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var requests []BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(requests) == 0 {
		http.Error(w, "Empty batch request", http.StatusBadRequest)
		return
	}

	responses := make([]BatchResponse, len(requests))

	for i, req := range requests {
		shortURL := randomid.GenerateRandomID()
		if err := store.Save(shortURL, req.OriginalURL); err != nil {
			http.Error(w, "Error saving URL", http.StatusInternalServerError)
			return
		}
		responses[i] = BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      baseURL + "/" + shortURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responses)
}
