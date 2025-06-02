package storage

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"os"
)

type Event struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type InMemory struct {
	data map[string]Event
}

type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	memory InMemory
}

func NewInMemory() *InMemory {
	return &InMemory{
		data: make(map[string]Event),
	}
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (f *FileStorage) WriteEvent(event *Event, db *sql.DB) error {
	_, err := db.Exec(
		`INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING`,
		event.UUID, event.ShortURL, event.OriginalURL,
	)
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if _, err := f.writer.Write(data); err != nil {
		return err
	}

	if err := f.writer.WriteByte('\n'); err != nil {
		return err
	}

	return f.writer.Flush()
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadEvents() (*InMemory, error) {
	if _, err := c.file.Seek(0, 0); err != nil {
		return nil, err
	}

	c.scanner = bufio.NewScanner(c.file)
	inMemory := NewInMemory()

	for c.scanner.Scan() {
		var event Event
		if err := json.Unmarshal(c.scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		inMemory.data[event.ShortURL] = event
	}

	if err := c.scanner.Err(); err != nil {
		return nil, err
	}

	return inMemory, nil
}

func (m *InMemory) Data() map[string]Event {
	return m.data
}

func (f *FileStorage) CloseFile() error {
	return f.file.Close()

}

func (c *Consumer) Close() error {
	return c.file.Close()
}
