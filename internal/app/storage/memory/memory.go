package memory

import (
	"errors"
	"sync"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
)

type MemoryStore struct {
	mu    sync.RWMutex
	data  map[string]string          // shortURL -> originalURL
	byUID map[string]map[string]bool // userID -> shortURL -> true
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data:  make(map[string]string),
		byUID: make(map[string]map[string]bool),
	}
}

func (m *MemoryStore) Save(shortURL, originalURL, userID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// сохраняем основную мапу short -> original
	m.data[shortURL] = originalURL

	// привязываем shortURL к userID
	if _, exists := m.byUID[userID]; !exists {
		m.byUID[userID] = make(map[string]bool)
	}
	m.byUID[userID][shortURL] = true

	return shortURL, nil
}

func (m *MemoryStore) Get(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, ok := m.data[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return originalURL, nil
}

func (m *MemoryStore) GetByOriginalURL(originalURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for short, orig := range m.data {
		if orig == originalURL {
			return short, nil
		}
	}
	return "", errors.New("original url not found")
}

func (m *MemoryStore) GetAllByUser(userID string) ([]dto.UserURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []dto.UserURL

	userShorts, ok := m.byUID[userID]
	if !ok {
		return result, nil
	}

	for shortURL := range userShorts {
		orig := m.data[shortURL]
		result = append(result, dto.UserURL{
			ShortURL:    shortURL,
			OriginalURL: orig,
		})
	}

	return result, nil
}

func (m *MemoryStore) BatchDelete(userID string, shortURLs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, shortURL := range shortURLs {
		delete(m.data, shortURL)
		if urls, ok := m.byUID[userID]; ok {
			delete(urls, shortURL)
		}
	}
	return nil
}
