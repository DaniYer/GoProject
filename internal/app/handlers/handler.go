package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

// NewHandleShortenURLv13 godoc
// @Summary      Создать короткую ссылку (JSON API v1.3)
// @Description  Принимает оригинальный URL в формате JSON и возвращает короткую ссылку.
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        request body dto.ShortenRequest true "Данные для сокращения"
// @Success      201 {object} dto.ShortenResponse "Короткая ссылка создана"
// @Success      409 {object} dto.ShortenResponse "Ссылка уже существует"
// @Failure      400 {string} string "invalid request"
// @Failure      500 {string} string "internal error"
// @Router       /api/shorten [post]
func NewHandleShortenURLv13(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)

		var req dto.ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		shortID, existed, err := svc.Shorten(req.URL, userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		resp := dto.ShortenResponse{Result: svc.BaseURL + "/" + shortID}
		w.Header().Set("Content-Type", "application/json")
		if existed {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		json.NewEncoder(w).Encode(resp)
	}
}
