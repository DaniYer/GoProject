package main

import (
	"encoding/json"
	"os"
	"sync"
)

type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	mu   sync.RWMutex
	data map[string]string
	file string
}

// NewStorage создает экземпляр хранилища, загружает данные из файла
func NewStorage(filePath string) (*Storage, error) {
	s := &Storage{
		data: make(map[string]string),
		file: filePath,
	}

	// Загружаем данные из файла, если он существует
	if err := s.loadFromFile(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

// loadFromFile загружает данные из JSON-файла
func (s *Storage) loadFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.file)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var urlList []URLData

	if err := decoder.Decode(&urlList); err != nil {
		return err
	}

	// Заполняем хранилище
	for _, entry := range urlList {
		s.data[entry.ShortURL] = entry.OriginalURL
	}
	return nil
}

// saveToFile сохраняет данные в JSON-файл
func (s *Storage) saveToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.file)
	if err != nil {
		return err
	}
	defer file.Close()

	var urlList []URLData
	for short, original := range s.data {
		urlList = append(urlList, URLData{
			ShortURL:    short,
			OriginalURL: original,
		})
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(urlList)
}

// SaveURL сохраняет URL в хранилище и записывает в файл
func (s *Storage) SaveURL(short, original string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[short] = original
	return s.saveToFile()
}

// GetURL возвращает оригинальный URL по сокращенному
func (s *Storage) GetURL(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exists := s.data[short]
	return url, exists
}
