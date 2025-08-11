package service

import (
	"sync"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
)

// URLService — бизнес-логика сервиса сокращения URL.
// Работает поверх хранилища (URLStore) и поддерживает:
//   - генерацию коротких ссылок (одиночную и пакетную);
//   - получение ссылок пользователя;
//   - асинхронное удаление ссылок через worker.
type URLService struct {
	Store   URLStore // интерфейс для работы с хранилищем
	BaseURL string   // базовый адрес для формирования полной короткой ссылки

	deleteQueue chan deleteTask
	once        sync.Once
}

// deleteTask — внутренний тип для отложенного удаления ссылок.
type deleteTask struct {
	userID    string
	shortURLs []string
}

// NewURLService создаёт и инициализирует новый сервис URL.
// Запускает горутину worker для асинхронного удаления ссылок.
func NewURLService(store URLStore, baseURL string) *URLService {
	svc := &URLService{
		Store:       store,
		BaseURL:     baseURL,
		deleteQueue: make(chan deleteTask, 1024),
	}

	// Worker запускается только один раз
	svc.once.Do(func() {
		go svc.deleteWorker()
	})

	return svc
}

// URLStore — контракт хранилища URL, реализуемый БД, файловым или in-memory хранилищем.
type URLStore interface {
	Save(shortURL, originalURL, userID string) (string, error)
	Get(shortURL string) (string, error)
	GetByOriginalURL(originalURL string) (string, error)
	GetAllByUser(userID string) ([]dto.UserURL, error)
	BatchDelete(userID string, shortURLs []string) error
}

// Shorten создаёт сокращённую ссылку для originalURL.
// Если ссылка уже существует, возвращает существующий shortURL и флаг duplicate=true.
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

// ShortenBatch обрабатывает пакетное сокращение ссылок.
// Возвращает массив с корреляционными ID и готовыми короткими URL.
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

// GetAllUserURLs возвращает все ссылки, сохранённые конкретным пользователем.
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

// Get возвращает оригинальный URL по сокращённому идентификатору.
func (s *URLService) Get(shortURL string) (string, error) {
	return s.Store.Get(shortURL)
}

// EnqueueURLsForDeletion добавляет ссылки в очередь на удаление.
// Удаление происходит асинхронно worker'ом.
func (s *URLService) EnqueueURLsForDeletion(userID string, shortURLs []string) {
	s.deleteQueue <- deleteTask{
		userID:    userID,
		shortURLs: shortURLs,
	}
}

// deleteWorker обрабатывает очередь на удаление, группируя задачи по userID.
func (s *URLService) deleteWorker() {
	batchSize := 100
	buffer := make([]deleteTask, 0, batchSize)

	for {
		task, ok := <-s.deleteQueue
		if !ok {
			break
		}

		buffer = append(buffer, task)

		if len(buffer) >= batchSize {
			s.flushBatch(buffer)
			buffer = buffer[:0]
		}
	}

	if len(buffer) > 0 {
		s.flushBatch(buffer)
	}
}

// flushBatch удаляет ссылки группами по userID.
func (s *URLService) flushBatch(tasks []deleteTask) {
	grouped := make(map[string][]string)

	for _, task := range tasks {
		grouped[task.userID] = append(grouped[task.userID], task.shortURLs...)
	}

	for userID, urls := range grouped {
		_ = s.Store.BatchDelete(userID, urls)
	}
}
