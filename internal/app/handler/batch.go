package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/utils"
)

func (h *Handler) BatchShortenHandler(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r.Context())

	var batch []dto.BatchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.service.SaveBatchURLs(context.Background(), userID, batch, h.config.BaseURL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}
