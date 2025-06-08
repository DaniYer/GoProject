package service

import (
	"sync"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
)

type URLService struct {
	Store   URLStore
	BaseURL string

	deleteQueue chan deleteTask
	once        sync.Once
}

type deleteTask struct {
	userID    string
	shortURLs []string
}

func NewURLService(store URLStore, baseURL string) *URLService {
	svc := &URLService{
		Store:       store,
		BaseURL:     baseURL,
		deleteQueue: make(chan deleteTask, 1024),
	}

	// Запускаем worker для асинхронного удаления
	svc.once.Do(func() {
		go svc.deleteWorker()
	})

	return svc
}

type URLStore interface {
	Save(shortURL, originalURL, userID string) (string, error)
	Get(shortURL string) (string, error)
	GetByOriginalURL(originalURL string) (string, error)
	GetAllByUser(userID string) ([]dto.UserURL, error)
	BatchDelete(userID string, shortURLs []string) error
}

func (s *URLService) Shorten(originalURL, userID string) (string, bool, error) {
	existingShortURL, err := s.Store.GetByOriginalURL(originalURL)
	if err == nil {
		return existingShortURL, true, nil
	}

	shortID := GenerateRandomID()
	shortID, err = s.Store.Save(shortID, originalURL, userID)
	if err != nil {
		return "", false, err
	}
	return shortID, false, nil
}

func (s *URLService) ShortenBatch(requests []dto.BatchRequest, userID string) ([]dto.BatchResponse, error) {
	responses := make([]dto.BatchResponse, len(requests))

	for i, req := range requests {
		shortURL := GenerateRandomID()
		if _, err := s.Store.Save(shortURL, req.OriginalURL, userID); err != nil {
			return nil, err
		}
		responses[i] = dto.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      s.BaseURL + "/" + shortURL,
		}
	}
	return responses, nil
}

func (s *URLService) GetAllUserURLs(userID string) ([]dto.UserURL, error) {
	urls, err := s.Store.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}
	for i := range urls {
		urls[i].ShortURL = s.BaseURL + "/" + urls[i].ShortURL
	}
	return urls, nil
}

func (s *URLService) Get(shortURL string) (string, error) {
	return s.Store.Get(shortURL)
}

// Асинхронная постановка задач на удаление
func (s *URLService) EnqueueURLsForDeletion(userID string, shortURLs []string) {
	s.deleteQueue <- deleteTask{
		userID:    userID,
		shortURLs: shortURLs,
	}
}

// Worker для асинхронной обработки удаления (fan-in)
func (s *URLService) deleteWorker() {
	batchSize := 100
	buffer := make([]deleteTask, 0, batchSize)

	for {
		task, ok := <-s.deleteQueue
		if !ok {
			break
		}

		buffer = append(buffer, task)

		// Если буфер наполнился — обрабатываем сразу
		if len(buffer) >= batchSize {
			s.flushBatch(buffer)
			buffer = buffer[:0]
		}
	}

	// Обрабатываем остаток при закрытии канала
	if len(buffer) > 0 {
		s.flushBatch(buffer)
	}
}

func (s *URLService) flushBatch(tasks []deleteTask) {
	// Собираем в Map по userID → список url
	grouped := make(map[string][]string)

	for _, task := range tasks {
		grouped[task.userID] = append(grouped[task.userID], task.shortURLs...)
	}

	for userID, urls := range grouped {
		_ = s.Store.BatchDelete(userID, urls)
	}
}
