package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// Record представляет запись для хранения URL.
type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// FileStore – реализация URLStore, использующая файл для персистентности.
type FileStore struct {
	filePath string
	data     map[string]string
}

func NewFileStore(filePath string, sugar *zap.SugaredLogger) (*FileStore, error) {
	store, err := loadStorageFromFile(filePath, sugar)
	if err != nil {
		return nil, err
	}
	return &FileStore{filePath: filePath, data: store}, nil
}

func (fs *FileStore) Save(shortURL, originalURL string) (string, error) {
	fs.data[shortURL] = originalURL
	rec := Record{
		UUID:        "", // для FileStore не требуется генерация UUID
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	return shortURL, appendRecordToFile(rec, fs.filePath)
}

func (fs *FileStore) Get(shortURL string) (string, error) {
	if url, ok := fs.data[shortURL]; ok {
		return url, nil
	}
	return "", fmt.Errorf("not found")
}

// loadStorageFromFile загружает записи из файла в карту.
func loadStorageFromFile(filePath string, sugar *zap.SugaredLogger) (map[string]string, error) {
	store := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var rec Record
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			sugar.Errorf("Ошибка парсинга записи: %v", err)
			continue
		}
		store[rec.ShortURL] = rec.OriginalURL
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return store, nil
}

// appendRecordToFile дописывает запись в файл.
func appendRecordToFile(rec Record, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	recBytes, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = file.Write(append(recBytes, '\n'))
	return err
}

func (fs *FileStore) SaveWithConflict(shortURL, originalURL string) (string, error) {
	shortURL, err := fs.Save(shortURL, originalURL)
	return shortURL, err
}
func (fs *FileStore) GetByOriginalURL(originalURL string) (string, error) {
	for short, orig := range fs.data {
		if orig == originalURL {
			return short, nil
		}
	}
	return "", fmt.Errorf("not found")
}
