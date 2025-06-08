package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

func NewHandleShortenURLv13(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleShortenURLv13(w, r, svc)
	}
}

func HandleShortenURLv13(w http.ResponseWriter, r *http.Request, svc *service.URLService) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	var req dto.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	shortID, isDuplicate, err := svc.ShortenJSON(req.URL, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := dto.ShortenResponse{
		Result: fmt.Sprintf("%s/%s", svc.BaseURL, shortID),
	}

	if isDuplicate {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
