package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/DaniYer/GoProject.git/internal/app/worker"
)

// NewBatchDeleteHandler godoc
// @Summary      Удалить сокращённые ссылки пачкой
// @Description  Принимает список коротких ссылок пользователя и отправляет их на асинхронное удаление
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        input body dto.DeleteRequest true "Список коротких ссылок для удаления"
// @Success      202 {string} string "Запрос принят на обработку"
// @Failure      400 {string} string "Некорректный запрос"
// @Router       /api/user/urls [delete]
func NewBatchDeleteHandler(svc *service.URLService, pool *worker.DeleteWorkerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)

		var req dto.DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		for _, shortURL := range req {
			task := worker.DeleteTask{
				UserID: userID,
				Short:  shortURL,
			}
			pool.AddTask(task)
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
