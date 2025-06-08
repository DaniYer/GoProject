package memory

import (
	"errors"
	"sync"
)

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]string),
	}
}

func (m *MemoryStore) Save(shortURL, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[shortURL] = originalURL
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
