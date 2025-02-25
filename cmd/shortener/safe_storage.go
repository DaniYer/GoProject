package main

import "sync"

// SafeStorage оборачивает карту с потокобезопасным доступом.
type SafeStorage struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewSafeStorage возвращает инициализированное хранилище.
func NewSafeStorage() *SafeStorage {
	return &SafeStorage{
		m: make(map[string]string),
	}
}

// Get безопасно возвращает значение по ключу.
func (s *SafeStorage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.m[key]
	return value, ok
}

// Set безопасно устанавливает значение по ключу.
func (s *SafeStorage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

// Delete безопасно удаляет значение по ключу.
func (s *SafeStorage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}

// Keys возвращает список всех ключей.
func (s *SafeStorage) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []string{}
	for key := range s.m {
		result = append(result, key)
	}
	return result
}
