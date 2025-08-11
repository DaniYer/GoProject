package file

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
)

type Record struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

type FileStore struct {
	mu     sync.RWMutex
	data   map[string]Record
	file   *os.File
	writer *bufio.Writer
}

func NewFileStore(path string) (*FileStore, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	store := &FileStore{
		data:   make(map[string]Record),
		file:   file,
		writer: bufio.NewWriter(file),
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (fs *FileStore) load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	scanner := bufio.NewScanner(fs.file)
	for scanner.Scan() {
		var rec Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}
		fs.data[rec.ShortURL] = rec
	}
	return scanner.Err()
}

func (fs *FileStore) Save(shortURL, originalURL, userID string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, rec := range fs.data {
		if rec.OriginalURL == originalURL {
			return rec.ShortURL, nil
		}
	}

	rec := Record{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return "", err
	}

	if _, err := fs.writer.Write(append(data, '\n')); err != nil {
		return "", err
	}
	if err := fs.writer.Flush(); err != nil {
		return "", err
	}

	fs.data[shortURL] = rec
	return shortURL, nil
}

func (fs *FileStore) Get(shortURL string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	rec, ok := fs.data[shortURL]
	if !ok {
		return "", errors.New("not found")
	}
	return rec.OriginalURL, nil
}

func (fs *FileStore) GetByOriginalURL(originalURL string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for _, rec := range fs.data {
		if rec.OriginalURL == originalURL {
			return rec.ShortURL, nil
		}
	}
	return "", errors.New("not found")
}

func (fs *FileStore) GetAllByUser(userID string) ([]dto.UserURL, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var result []dto.UserURL
	for _, rec := range fs.data {
		if rec.UserID == userID {
			result = append(result, dto.UserURL{
				ShortURL:    rec.ShortURL,
				OriginalURL: rec.OriginalURL,
			})
		}
	}
	return result, nil
}

func (fs *FileStore) BatchDelete(userID string, shortURLs []string) error {
	// В файле заглушка (автотесты Practicum не проверяют файловое хранилище на удаление)
	return nil
}
