package memory

import "fmt"

// MemoryStore – реализация URLStore с использованием in-memory карты.
type MemoryStore struct {
	data map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]string)}
}

func (ms *MemoryStore) Save(shortURL, originalURL string) error {
	ms.data[shortURL] = originalURL
	return nil
}

func (ms *MemoryStore) Get(shortURL string) (string, error) {
	if url, ok := ms.data[shortURL]; ok {
		return url, nil
	}
	return "", fmt.Errorf("not found")
}
func (ms *MemoryStore) SaveWithConflict(shortURL, originalURL string) (string, error) {
	err := ms.Save(shortURL, originalURL)
	return shortURL, err
}
