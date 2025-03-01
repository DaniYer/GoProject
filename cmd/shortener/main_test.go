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

	// Создаём временный файл для тестов файлового хранилища
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	if err != nil {
		panic(err)
	}
	storageFilePath := tmpFile.Name()
	tmpFile.Close()
	// Устанавливаем переменную окружения для использования файлового хранилища
	os.Setenv("FILE_STORAGE_PATH", storageFilePath)

	code := m.Run()
	os.Remove(storageFilePath)
	os.Exit(code)
}

func TestShortenedURL(t *testing.T) {
	store := NewMemoryStore() // используем in-memory хранилище для теста
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
	w := httptest.NewRecorder()
	baseURL := "http://localhost:8080"
	shortenedURL(w, req, baseURL, store)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("ожидался статус 201, получен %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	responseStr := string(body)
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
	if original, err := store.Get(shortURL); err != nil || original != "http://example.com" {
		t.Errorf("хранилище не обновлено корректно, ожидался http://example.com, получено %v", original)
	}
}

func TestRedirectedURL(t *testing.T) {
	store := NewMemoryStore()
	shortURL := "abcdefgh"
	originalURL := "http://example.org"
	store.Save(shortURL, originalURL)

	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
	w := httptest.NewRecorder()
	redirectedURL(w, req, store)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("ожидался статус 307, получен %d", res.StatusCode)
	}
	loc := res.Header.Get("Location")
	if loc != originalURL {
		t.Errorf("ожидался редирект на %s, получен %s", originalURL, loc)
	}
}

func TestJSONHandler(t *testing.T) {
	store := NewMemoryStore()
	jsonPayload := `{"url": "http://example.net"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	jsonHandler(w, req, store)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("ожидался статус 201, получен %d", res.StatusCode)
	}
	var resp redirectURL
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка декодирования JSON ответа: %v", err)
	}
	expectedPrefix := "http://localhost:8080/"
	if !strings.HasPrefix(resp.Result, expectedPrefix) {
		t.Errorf("ожидался префикс %s, получено %s", expectedPrefix, resp.Result)
	}
	parts := strings.Split(resp.Result, "/")
	if len(parts) < 2 {
		t.Errorf("неверный формат результата: %s", resp.Result)
	}
	shortURL := parts[len(parts)-1]
	if storedURL, err := store.Get(shortURL); err != nil || storedURL != "http://example.net" {
		t.Errorf("хранилище не обновлено корректно, ожидался http://example.net, получено %v", storedURL)
	}
}

func TestPersistence(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "storage_test_*.json")
	if err != nil {
		t.Fatalf("не удалось создать временный файл: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Создаём FileStore
	fs, err := NewFileStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("Ошибка инициализации FileStore: %v", err)
	}

	// Сохраняем запись
	err = fs.Save("testurl1", "http://test1.com")
	if err != nil {
		t.Fatalf("ошибка Save в FileStore: %v", err)
	}

	// Создаём новый FileStore для проверки загрузки данных из файла
	fs2, err := NewFileStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("Ошибка инициализации FileStore: %v", err)
	}
	if original, err := fs2.Get("testurl1"); err != nil || original != "http://test1.com" {
		t.Errorf("тест персистентности не пройден, ожидался http://test1.com, получено %v", original)
	}
}

func TestPingHandler_Error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	pingHandler(w, req, "invalid-dsn")
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Ожидался статус 500, получен %d", res.StatusCode)
	}
}

func TestPingHandler_Success(t *testing.T) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" || dsn == "localDB" {
		t.Skip("Пропуск теста, так как не настроена корректная строка подключения к БД")
	}
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	pingHandler(w, req, dsn)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус 200, получен %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "Связь налажена") {
		t.Errorf("Неверное тело ответа: %s", string(body))
	}
}
