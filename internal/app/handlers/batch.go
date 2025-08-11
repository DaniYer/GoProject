package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

// NewBatchShortenURLHandler godoc
// @Summary      Сократить ссылки пачкой
// @Description  Принимает массив исходных URL и возвращает массив сокращённых ссылок
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        input body []dto.BatchRequest true "Список ссылок для сокращения"
// @Success      201 {array} dto.BatchResponse
// @Failure      400 {string} string "Некорректный запрос"
// @Failure      500 {string} string "Внутренняя ошибка"
// @Router       /api/shorten/batch [post]
func NewBatchShortenURLHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)

		var req []dto.BatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		responses, err := svc.ShortenBatch(req, userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responses)
	}
}
