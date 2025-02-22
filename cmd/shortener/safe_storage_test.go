package main

import (
	"sync"
	"testing"
)

func TestSafeStorage(t *testing.T) {
	ss := NewSafeStorage()
	key := "testKey"
	value := "testValue"

	// Тестируем Set и Get
	ss.Set(key, value)
	if v, ok := ss.Get(key); !ok || v != value {
		t.Errorf("Ожидалось значение %s, получено %s", value, v)
	}

	// Тестируем Delete
	ss.Delete(key)
	if _, ok := ss.Get(key); ok {
		t.Error("Ключ не был удалён")
	}
}

func TestSafeStorageConcurrency(t *testing.T) {
	ss := NewSafeStorage()
	const goroutines = 100
	var wg sync.WaitGroup

	// Запускаем несколько горутин, которые одновременно записывают значения
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + string(i)
			ss.Set(key, "value")
		}(i)
	}
	wg.Wait()

	// Проверяем, что все записи присутствуют
	keys := ss.Keys()
	if len(keys) != goroutines {
		t.Errorf("Ожидалось %d ключей, получено %d", goroutines, len(keys))
	}
}
