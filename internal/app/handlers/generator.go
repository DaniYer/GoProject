package handlers

import (
	"io"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

func NewGenerateShortURLHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		GenerateShortURLHandler(w, r, svc)
	}
}

func GenerateShortURLHandler(w http.ResponseWriter, r *http.Request, svc *service.URLService) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)

	existingShortURL, err := svc.Store.GetByOriginalURL(originalURL)
	resultURL := svc.BaseURL + "/" + existingShortURL

	if err == nil {
		w.WriteHeader(http.StatusConflict)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resultURL))
		return
	}

	shortID := service.GenerateRandomID()
	shortID, err = svc.Store.Save(shortID, originalURL, userID)
	if err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resultURL = svc.BaseURL + "/" + shortID
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(resultURL))
}
