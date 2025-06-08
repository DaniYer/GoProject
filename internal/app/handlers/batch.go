package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

func NewBatchShortenURLHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)

		var req []dto.BatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		responses := make([]dto.BatchResponse, 0, len(req))
		for _, item := range req {
			shortID, _, err := svc.Shorten(item.OriginalURL, userID)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			responses = append(responses, dto.BatchResponse{
				CorrelationID: item.CorrelationID,
				ShortURL:      svc.BaseURL + "/" + shortID,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	}
}
