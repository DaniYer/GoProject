package worker

import (
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

type DeleteTask struct {
	UserID    string
	ShortURLs []string
}

type DeleteWorker struct {
	Queue chan DeleteTask
	Store service.URLStore
}

func NewDeleteWorker(store service.URLStore, queueSize int) *DeleteWorker {
	return &DeleteWorker{
		Queue: make(chan DeleteTask, queueSize),
		Store: store,
	}
}

func (w *DeleteWorker) Start() {
	go func() {
		for task := range w.Queue {
			_ = w.Store.BatchDelete(task.UserID, task.ShortURLs)
		}
	}()
}
