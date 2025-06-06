package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/google/uuid"
)

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	filePath string
	mu       sync.RWMutex
	urls     map[string]dto.UserURL
	user     map[string][]dto.UserURL
	counter  int
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		urls:     make(map[string]dto.UserURL),
		user:     make(map[string][]dto.UserURL),
		counter:  0,
	}

	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (fs *FileStorage) SaveURL(ctx context.Context, userID, originalURL string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	shortID := uuid.New().String()[:8]
	fs.counter++

	url := dto.UserURL{
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	fs.urls[shortID] = url
	fs.user[userID] = append(fs.user[userID], url)

	if err := fs.appendRecord(shortID, originalURL); err != nil {
		return "", err
	}

	return shortID, nil
}

func (fs *FileStorage) GetOriginalURL(ctx context.Context, shortID string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	url, ok := fs.urls[shortID]
	if !ok {
		return "", nil
	}
	return url.OriginalURL, nil
}

func (fs *FileStorage) GetUserURLs(ctx context.Context, userID string) ([]dto.UserURL, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	urls := fs.user[userID]
	return urls, nil
}

func (fs *FileStorage) appendRecord(shortID, originalURL string) error {
	file, err := os.OpenFile(fs.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	rec := record{
		UUID:        strconv.Itoa(fs.counter),
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(rec); err != nil {
		return err
	}
	return nil
}

func (fs *FileStorage) load() error {
	file, err := os.Open(fs.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var rec record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return err
		}
		url := dto.UserURL{
			ShortURL:    rec.ShortURL,
			OriginalURL: rec.OriginalURL,
		}
		fs.urls[rec.ShortURL] = url
		fs.user[rec.UUID] = append(fs.user[rec.UUID], url)

		id, _ := strconv.Atoi(rec.UUID)
		if id > fs.counter {
			fs.counter = id
		}
	}

	return scanner.Err()
}

// SaveBatchURLs saves a batch of URLs for a user and returns the corresponding short URLs.
// It generates a unique short ID for each URL and appends it to the user's list of URLs.
func (fs *FileStorage) SaveBatchURLs(ctx context.Context, userID string, batch []dto.BatchRequestItem, baseURL string) ([]dto.BatchResponseItem, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	var result []dto.BatchResponseItem

	for _, item := range batch {
		fs.counter++
		shortID := uuid.New().String()[:8]

		url := dto.UserURL{
			ShortURL:    shortID,
			OriginalURL: item.OriginalURL,
		}

		fs.urls[shortID] = url
		fs.user[userID] = append(fs.user[userID], url)

		if err := fs.appendRecord(shortID, item.OriginalURL); err != nil {
			return nil, err
		}

		result = append(result, dto.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      baseURL + "/" + shortID,
		})
	}

	return result, nil
}
