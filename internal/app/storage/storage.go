package storage

import (
	"bufio"
	"encoding/json"
	"os"
)

type Event struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteEvent(event *Event) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

type Consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый scanner
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadEvents() ([]Event, error) {
	// Сброс указателя файла в начало (если нужно, чтобы читать все данные с начала)
	if _, err := c.file.Seek(0, 0); err != nil {
		return nil, err
	}
	// Переинициализируем сканер, чтобы начать чтение заново
	c.scanner = bufio.NewScanner(c.file)

	var events []Event
	for c.scanner.Scan() {
		var event Event
		if err := json.Unmarshal(c.scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := c.scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

// CloseFile закрывает файл Producer-а (для тестирования ошибок)
func (p *Producer) CloseFile() error {
	return p.file.Close()
}
