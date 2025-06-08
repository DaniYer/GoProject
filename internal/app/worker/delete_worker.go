package worker

import (
	"sync"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/service"
)

// DeleteTask — задача на удаление
type DeleteTask struct {
	UserID string
	Short  string
}

type DeleteWorkerPool struct {
	store     service.URLStore
	tasks     chan DeleteTask
	wg        sync.WaitGroup
	workerNum int
	flushDur  time.Duration
}

func NewDeleteWorkerPool(store service.URLStore, bufferSize int, workerNum int, flushDur time.Duration) *DeleteWorkerPool {
	return &DeleteWorkerPool{
		store:     store,
		tasks:     make(chan DeleteTask, bufferSize),
		workerNum: workerNum,
		flushDur:  flushDur,
	}
}

func (p *DeleteWorkerPool) Start() {
	for i := 0; i < p.workerNum; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *DeleteWorkerPool) AddTask(task DeleteTask) {
	p.tasks <- task
}

func (p *DeleteWorkerPool) Shutdown() {
	close(p.tasks)
	p.wg.Wait()
}

func (p *DeleteWorkerPool) worker() {
	defer p.wg.Done()

	batch := make(map[string][]string)
	ticker := time.NewTicker(p.flushDur)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				p.flush(batch)
				return
			}
			batch[task.UserID] = append(batch[task.UserID], task.Short)
			if len(batch[task.UserID]) >= 100 {
				p.flushUser(task.UserID, batch)
			}
		case <-ticker.C:
			p.flush(batch)
		}
	}
}

func (p *DeleteWorkerPool) flush(batch map[string][]string) {
	for userID, urls := range batch {
		p.store.BatchDelete(userID, urls)
		delete(batch, userID)
	}
}

func (p *DeleteWorkerPool) flushUser(userID string, batch map[string][]string) {
	p.store.BatchDelete(userID, batch[userID])
	delete(batch, userID)
}
