package memory

import (
	"errors"
	"sync"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
)

type StoredURL struct {
	OriginalURL string
	UserID      string
	Deleted     bool
}

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]StoredURL
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]StoredURL),
	}
}

func (m *MemoryStore) Save(shortURL, originalURL, userID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Проверяем существование по OriginalURL
	for short, record := range m.data {
		if record.OriginalURL == originalURL && !record.Deleted {
			return short, nil
		}
	}

	m.data[shortURL] = StoredURL{
		OriginalURL: originalURL,
		UserID:      userID,
		Deleted:     false,
	}
	return shortURL, nil
}

func (m *MemoryStore) Get(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.data[shortURL]
	if !ok {
		return "", errors.New("not found")
	}
	if record.Deleted {
		return "", errors.New("gone")
	}
	return record.OriginalURL, nil
}

func (m *MemoryStore) GetByOriginalURL(originalURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for short, record := range m.data {
		if record.OriginalURL == originalURL && !record.Deleted {
			return short, nil
		}
	}
	return "", errors.New("not found")
}

func (m *MemoryStore) GetAllByUser(userID string) ([]dto.UserURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []dto.UserURL
	for short, record := range m.data {
		if record.UserID == userID && !record.Deleted {
			result = append(result, dto.UserURL{
				ShortURL:    short,
				OriginalURL: record.OriginalURL,
			})
		}
	}
	return result, nil
}

func (m *MemoryStore) BatchDelete(userID string, shortURLs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, shortURL := range shortURLs {
		record, ok := m.data[shortURL]
		if ok && record.UserID == userID {
			record.Deleted = true
			m.data[shortURL] = record
		}
	}
	return nil
}
