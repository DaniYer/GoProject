package shortener

import (
	"io"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/randomid"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

func NewGenerateShortURLHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		GenerateShortURLHandler(w, r, svc)
	}
}

func GenerateShortURLHandler(w http.ResponseWriter, r *http.Request, svc *service.URLService) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)

	existingShortURL, err := svc.Store.GetByOriginalURL(originalURL)
	if err == nil {
		resultURL := svc.BaseURL + "/" + existingShortURL
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(resultURL))
		return
	}

	shortID := randomid.GenerateRandomID()

	shortID, err = svc.Store.Save(shortID, originalURL)
	if err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resultURL := svc.BaseURL + "/" + shortID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

// GenerateShortURLHandler обрабатывает запросы на генерацию коротких URL
