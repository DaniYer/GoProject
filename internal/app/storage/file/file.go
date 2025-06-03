package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/DaniYer/GoProject.git/internal/app/storage"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

// FileStore – реализация URLStore, использующая файл для персистентности.
type FileStore struct {
	filePath string
	data     map[string]string
}

func NewFileStore(filePath string) (*FileStore, error) {
	store, err := loadStorageFromFile(filePath)
	if err != nil {
		return nil, err
	}
	return &FileStore{filePath: filePath, data: store}, nil
}

func (fs *FileStore) Save(shortURL, originalURL string) error {
	fs.data[shortURL] = originalURL
	rec := storage.Record{
		UUID:        "", // для FileStore не требуется генерация UUID
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	return appendRecordToFile(rec, fs.filePath)
}

func (fs *FileStore) Get(shortURL string) (string, error) {
	if url, ok := fs.data[shortURL]; ok {
		return url, nil
	}
	return "", fmt.Errorf("not found")
}

// loadStorageFromFile загружает записи из файла в карту.
func loadStorageFromFile(filePath string) (map[string]string, error) {
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
		var rec storage.Record
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
func appendRecordToFile(rec storage.Record, filePath string) error {
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
