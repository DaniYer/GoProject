package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/DaniYer/GoProject.git/internal/app/worker"
)

type BatchDeleteDeps struct {
	Service *service.URLService
	Worker  *worker.DeleteWorker
}

func NewBatchDeleteHandler(deps BatchDeleteDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)

		var req dto.DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Передаем в воркер
		deps.Worker.Queue <- worker.DeleteTask{UserID: userID, ShortURLs: req}

		w.WriteHeader(http.StatusAccepted)
	}
}
