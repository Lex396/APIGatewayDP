package storage

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err := New(ctx, "postgres://filteruser:fpassword@localhost:5432/comments?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStore_AddComment(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	dataBase, err := New(ctx, "postgres://filteruser:fpassword@localhost:5432/comments?sslmode=disable")
	comment := Comment{
		NewsID:  2,
		Content: "Текст проверки",
	}
	dataBase.AddComment(comment)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Создана запись.")
}

func TestStore_DeleteComment(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	dataBase, err := New(ctx, "postgres://filteruser:fpassword@localhost:5432/comments?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	err = dataBase.DeleteComment(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Удалена запись.")
}
