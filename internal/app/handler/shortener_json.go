package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/repository"
	"github.com/DaniYer/GoProject.git/internal/app/utils"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

func (h *Handler) ShortenJSONHandler(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r.Context())

	var input struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortID, err := h.service.SaveURL(context.Background(), userID, input.URL)
	if err != nil {
		if err == repository.ErrDuplicateURL {
			resp := map[string]string{"result": h.config.BaseURL + "/" + shortID}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"result": h.config.BaseURL + "/" + shortID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
