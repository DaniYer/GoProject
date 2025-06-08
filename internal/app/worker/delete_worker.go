package worker

import (
	"sync"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/service"
)

type DeleteTask struct {
	UserID string
	Short  string
}

type DeleteWorkerPool struct {
	service  *service.URLService
	tasks    chan DeleteTask
	wg       sync.WaitGroup
	flushDur time.Duration
	batch    map[string][]string
	mu       sync.Mutex
}

func NewDeleteWorkerPool(service *service.URLService, bufferSize int, flushDur time.Duration) *DeleteWorkerPool {
	return &DeleteWorkerPool{
		service:  service,
		tasks:    make(chan DeleteTask, bufferSize),
		flushDur: flushDur,
		batch:    make(map[string][]string),
	}
}

func (p *DeleteWorkerPool) Start() {
	p.wg.Add(1)
	go p.worker()
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
	ticker := time.NewTicker(p.flushDur)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				p.flush()
				return
			}
			p.mu.Lock()
			p.batch[task.UserID] = append(p.batch[task.UserID], task.Short)
			p.mu.Unlock()
		case <-ticker.C:
			p.flush()
		}
	}
}

func (p *DeleteWorkerPool) flush() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for userID, urls := range p.batch {
		_ = p.service.Store.BatchDelete(userID, urls)
	}
	p.batch = make(map[string][]string)
}
