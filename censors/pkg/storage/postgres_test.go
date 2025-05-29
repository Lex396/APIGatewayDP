package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

var connStr = "postgres://filteruser:fpassword@localhost:5432/comments?sslmode=disable"

func TestMain(m *testing.M) {
	// Проверка соединения один раз
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := New(ctx, connStr)
	if err != nil {
		panic("Не удалось подключиться к БД: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestStore_AddList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store, err := New(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}

	err = store.AddList(Stop{StopList: "ups"})
	if err != nil {
		t.Errorf("Ошибка при добавлении: %v", err)
	}
}

func TestStore_AllList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store, err := New(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}

	list, err := store.AllList()
	if err != nil {
		t.Fatalf("Ошибка при чтении списка: %v", err)
	}

	if len(list) == 0 {
		t.Error("Список стоп-слов пуст")
	} else {
		t.Logf("Найдено %d записей", len(list))
	}
}
