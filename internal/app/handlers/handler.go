package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

func NewHandleShortenURLv13(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)

		var req dto.ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		shortID, existed, err := svc.ShortenJSON(req.URL, userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		status := http.StatusCreated
		if existed {
			status = http.StatusConflict
		}

		resp := dto.ShortenResponse{Result: svc.BaseURL + "/" + shortID}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(resp)
	}
}
