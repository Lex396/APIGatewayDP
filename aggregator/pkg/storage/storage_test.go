package storage

import (
	"APIGateway/aggregator/pkg/logger"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func initMemDB(t *testing.T) *MemDB {
	db := NewMemDB()
	logInstance, err := logger.NewLogger("test.log")
	if err != nil {
		t.Fatalf("Ошибка создания логгера: %v", err)
	}
	_ = logInstance
	return db
}

func TestMemDB_SavePost(t *testing.T) {
	db := initMemDB(t)
	now := time.Now()

	post := Post{
		Title:   "Test Title",
		Content: "Test Content",
		PubTime: now,
		Link:    "http://example.com",
	}

	err := db.SavePost(post, nil)
	if err != nil {
		t.Errorf("Ошибка при сохранении поста: %v", err)
	}

	posts, err := db.GetLastPosts(1, now)
	if err != nil {
		t.Errorf("Ошибка при получении постов: %v", err)
	}

	t.Logf("Полученные посты: %+v", posts)

	if len(posts) != 1 {
		t.Errorf("Ожидался 1 пост, получено %d", len(posts))
	}
}

func TestMemDB_GetLastPosts_EmptyDB(t *testing.T) {
	db := initMemDB(t)
	now := time.Now()

	posts, err := db.GetLastPosts(5, now)
	if err != nil {
		t.Errorf("Ошибка при получении постов из пустой БД: %v", err)
	}

	if len(posts) != 0 {
		t.Errorf("Ожидался пустой список, получено %d записей", len(posts))
	}
}

func TestMemDB_GetLastPosts_Limit(t *testing.T) {
	db := initMemDB(t)
	now := time.Now()

	for i := 0; i < 5; i++ {
		post := Post{
			ID:      i,
			Title:   fmt.Sprintf("Post %d", i),
			Content: "Content",
			PubTime: now.Add(time.Duration(-i) * time.Hour),
			Link:    fmt.Sprintf("http://example.com/%d", i),
		}
		err := db.SavePost(post, nil)
		if err != nil {
			t.Fatalf("Ошибка при сохранении поста: %v", err)
		}
	}

	posts, err := db.GetLastPosts(3, now)
	if err != nil {
		t.Fatalf("Ошибка при получении постов: %v", err)
	}

	if len(posts) != 3 {
		t.Fatalf("Ожидалось 3 поста, получено %d", len(posts))
	}
}

func TestLoadConfig(t *testing.T) {
	logInstance, _ := logger.NewLogger("test.log")

	tempFile, err := os.CreateTemp("", "config_*.json")
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tempFile.Name())

	configData := Config{
		RSS:           []string{"http://example.com/rss"},
		RequestPeriod: 10,
	}

	data, _ := json.Marshal(configData)
	tempFile.Write(data)
	tempFile.Close()

	loadedConfig, err := LoadConfig(tempFile.Name(), logInstance)
	if err != nil {
		t.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	if loadedConfig.RequestPeriod != 10 {
		t.Errorf("Ожидался RequestPeriod=10, получено %d", loadedConfig.RequestPeriod)
	}
	if len(loadedConfig.RSS) != 1 || loadedConfig.RSS[0] != "http://example.com/rss" {
		t.Errorf("Ошибка в RSS, получено: %+v", loadedConfig.RSS)
	}
}
