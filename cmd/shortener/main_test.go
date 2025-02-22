package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugar = logger.Sugar()

	// Создаём временный файл для хранения данных
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	if err != nil {
		panic(err)
	}
	storageFilePath = tmpFile.Name()
	tmpFile.Close()

	code := m.Run()

	// Очистка временного файла после тестов
	os.Remove(storageFilePath)
	os.Exit(code)
}

// TestShortenedURL проверяет обработчик POST "/" для создания сокращённого URL.
func TestShortenedURL(t *testing.T) {
	// Сбрасываем глобальное хранилище и счетчик
	storage = map[string]string{}
	nextUUID = 1

	// Формируем запрос с телом, содержащим исходный URL
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
	w := httptest.NewRecorder()
	baseURL := "http://localhost:8080"

	// Вызываем обработчик
	shortenedURL(w, req, baseURL)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("ожидался статус 201, получен %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	responseStr := string(body)

	// Проверяем, что ответ начинается с базового URL и содержит сокращённую часть длиной 8 символов
	if !strings.HasPrefix(responseStr, baseURL+"/") {
		t.Errorf("ответ не начинается с базового URL: %s", responseStr)
	}
	parts := strings.Split(responseStr, "/")
	if len(parts) < 2 {
		t.Errorf("неверный формат ответа: %s", responseStr)
	}
	shortURL := parts[len(parts)-1]
	if len(shortURL) != 8 {
		t.Errorf("ожидалась длина сокращённого URL равная 8, получено %d", len(shortURL))
	}
	// Проверяем, что в хранилище появился соответствующий URL
	if original, exists := storage[shortURL]; !exists || original != "http://example.com" {
		t.Errorf("хранилище не обновлено корректно, ожидался http://example.com, получено %v", original)
	}
}

// TestRedirectedURL проверяет корректность редиректа по сокращённому URL.
func TestRedirectedURL(t *testing.T) {
	// Инициализируем хранилище с заранее известным значением
	storage = map[string]string{}
	shortURL := "abcdefgh"
	originalURL := "http://example.org"
	storage[shortURL] = originalURL

	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
	w := httptest.NewRecorder()

	redirectedURL(w, req)
	res := w.Result()
	defer res.Body.Close()

	// Проверяем, что статус редиректа соответствует 307 (Temporary Redirect)
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("ожидался статус 307, получен %d", res.StatusCode)
	}
	// Проверяем заголовок Location
	loc := res.Header.Get("Location")
	if loc != originalURL {
		t.Errorf("ожидался редирект на %s, получен %s", originalURL, loc)
	}
}

// TestJSONHandler проверяет создание сокращённого URL через JSON API.
func TestJSONHandler(t *testing.T) {
	// Сбрасываем хранилище и счетчик
	storage = map[string]string{}
	nextUUID = 1

	// Подготавливаем JSON-запрос
	jsonPayload := `{"url": "http://example.net"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	jsonHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("ожидался статус 201, получен %d", res.StatusCode)
	}
	var resp redirectURL
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка декодирования JSON ответа: %v", err)
	}

	// Проверяем, что результат начинается с базового URL
	expectedPrefix := "http://localhost:8080/"
	if !strings.HasPrefix(resp.Result, expectedPrefix) {
		t.Errorf("ожидался префикс %s, получено %s", expectedPrefix, resp.Result)
	}
	// Извлекаем сокращённый URL и сверяем с хранилищем
	parts := strings.Split(resp.Result, "/")
	if len(parts) < 2 {
		t.Errorf("неверный формат результата: %s", resp.Result)
	}
	shortURL := parts[len(parts)-1]
	if storedURL, exists := storage[shortURL]; !exists || storedURL != "http://example.net" {
		t.Errorf("хранилище не обновлено корректно, ожидался http://example.net, получено %v", storedURL)
	}
}

// TestPersistence проверяет функции appendRecord и loadStorage, используя временный файл.
func TestPersistence(t *testing.T) {
	// Сбрасываем глобальное хранилище и счетчик
	storage = map[string]string{}
	nextUUID = 1

	// Создаем временный файл для хранения данных
	tmpFile, err := os.CreateTemp("", "storage_test_*.json")
	if err != nil {
		t.Fatalf("не удалось создать временный файл: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Устанавливаем путь к файлу в глобальную переменную
	storageFilePath = tmpFile.Name()

	// Создаем тестовую запись
	record := Record{
		UUID:        "1",
		ShortURL:    "testurl1",
		OriginalURL: "http://test1.com",
	}
	if err := appendRecord(record, storageFilePath); err != nil {
		t.Fatalf("ошибка appendRecord: %v", err)
	}

	// Сбрасываем хранилище и счётчик, чтобы проверить загрузку данных
	storage = map[string]string{}
	nextUUID = 1

	if err := loadStorage(storageFilePath); err != nil {
		t.Fatalf("ошибка loadStorage: %v", err)
	}
	// Проверяем, что запись загрузилась корректно
	if original, exists := storage["testurl1"]; !exists || original != "http://test1.com" {
		t.Errorf("тест персистентности не пройден, ожидался http://test1.com, получено %v", original)
	}
}
