package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func ConnectByURL(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	return db, nil
}
