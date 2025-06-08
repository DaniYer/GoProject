package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/DaniYer/GoProject.git/internal/app/worker"
)

type DeleteHandler struct {
	Svc  *service.URLService
	Pool *worker.DeleteWorkerPool
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	var req dto.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	for _, short := range req {
		h.Pool.AddTask(worker.DeleteTask{
			UserID: userID,
			Short:  short,
		})
	}

	w.WriteHeader(http.StatusAccepted)
}
