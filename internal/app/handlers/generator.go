package handlers

import (
	"encoding/json"
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
	resp := map[string]string{"result": svc.BaseURL + "/" + existingShortURL}

	w.Header().Set("Content-Type", "application/json")
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(resp)
		return
	}

	shortID := service.GenerateRandomID()
	shortID, err = svc.Store.Save(shortID, originalURL, userID)
	if err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resp = map[string]string{"result": svc.BaseURL + "/" + shortID}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
