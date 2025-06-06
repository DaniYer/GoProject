package repository

import (
	"context"
	"sync"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/google/uuid"
)

type InMemoryRepository struct {
	mu   sync.RWMutex
	urls map[string]dto.UserURL
	user map[string][]dto.UserURL
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		urls: make(map[string]dto.UserURL),
		user: make(map[string][]dto.UserURL),
	}
}

func (mem *InMemoryRepository) SaveURL(ctx context.Context, userID, originalURL string) (string, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	shortID := uuid.New().String()[:8]
	url := dto.UserURL{
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	mem.urls[shortID] = url
	mem.user[userID] = append(mem.user[userID], url)

	return shortID, nil
}

func (mem *InMemoryRepository) GetOriginalURL(ctx context.Context, shortID string) (string, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	url, ok := mem.urls[shortID]
	if !ok {
		return "", nil
	}
	return url.OriginalURL, nil
}

func (mem *InMemoryRepository) GetUserURLs(ctx context.Context, userID string) ([]dto.UserURL, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	urls := mem.user[userID]
	return urls, nil
}

// SaveBatchURLs saves a batch of URLs for a user and returns the corresponding short URLs.
// It generates a unique short ID for each URL and appends it to the user's list of URLs.
func (mem *InMemoryRepository) SaveBatchURLs(ctx context.Context, userID string, batch []dto.BatchRequestItem, baseURL string) ([]dto.BatchResponseItem, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	var result []dto.BatchResponseItem

	for _, item := range batch {
		shortID := uuid.New().String()[:8]
		url := dto.UserURL{
			ShortURL:    shortID,
			OriginalURL: item.OriginalURL,
		}
		mem.urls[shortID] = url
		mem.user[userID] = append(mem.user[userID], url)

		result = append(result, dto.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      baseURL + "/" + shortID,
		})
	}

	return result, nil
}
