package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	logInstance, err := NewLogger("test.log")
	if err != nil {
		t.Fatalf("Ошибка создания логгера: %v", err)
	}

	logInstance.InfoWithRequestID("test-id", "Тест INFO")
	logInstance.ErrorWithRequestID("test-id", "Тест ERROR")

	if _, err := os.Stat("test.log"); os.IsNotExist(err) {
		t.Errorf("Файл логов не создан")
	}
}
