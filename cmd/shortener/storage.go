package main

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

// Storage — структура для хранения URL
type Storage struct {
	mu       sync.RWMutex
	filePath string
	data     map[string]string
}

// NewStorage создаёт новое хранилище и загружает данные из файла
func NewStorage(filePath string) (*Storage, error) {
	storage := &Storage{
		filePath: filePath,
		data:     make(map[string]string),
	}

	// Загружаем данные, если файл существует
	if err := storage.load(); err != nil {
		return nil, err
	}

	return storage, nil
}

// SaveURL сохраняет URL в хранилище
func (s *Storage) SaveURL(short, original string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[short] = original
	return s.save()
}

// GetURL получает оригинальный URL по короткому идентификатору
func (s *Storage) GetURL(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exists := s.data[short]
	return url, exists
}

// save сохраняет данные в файл
func (s *Storage) save() error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(s.data)
}

// load загружает данные из файла
func (s *Storage) load() error {
	file, err := os.Open(s.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil // Файл не существует — начинаем с пустого хранилища
	}
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&s.data)
}
