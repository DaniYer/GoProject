package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/utils"
)

func (h *Handler) UserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r.Context())

	urls, err := h.service.GetUserURLs(context.Background(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(urls)
}
