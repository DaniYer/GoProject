package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

// GetUserURLsHandler godoc
// @Summary      Получить все сокращённые ссылки пользователя
// @Description  Возвращает список всех сокращённых ссылок, принадлежащих текущему пользователю
// @Tags         urls
// @Produce      json
// @Success      200  {array}  dto.UserURL   "Список ссылок"
// @Success      204  {string} string        "Нет ссылок"
// @Failure      500  {string} string        "Внутренняя ошибка сервера"
// @Router       /api/user/urls [get]
func GetUserURLsHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)

		urls, err := svc.GetAllUserURLs(userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(urls)
	}
}
