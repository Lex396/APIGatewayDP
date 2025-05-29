package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

// Store Хранилище данных
type Store struct {
	db *pgxpool.Pool
}

// New Конструктор объекта хранилища
func New(ctx context.Context, constr string) (*Store, error) {
	db, err := pgxpool.New(ctx, constr)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

// AllList Выводит все комментарии.
func (p *Store) AllList() ([]Stop, error) {
	rows, err := p.db.Query(context.Background(), "SELECT * FROM stop")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Stop
	for rows.Next() {
		var c Stop
		err = rows.Scan(&c.ID, &c.StopList)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}

	return list, rows.Err()
}

// AddList Добавляет комментарии в стоп лист.
func (p Store) AddList(c Stop) error {
	_, err := p.db.Exec(context.Background(),
		"INSERT INTO stop (stop_list) VALUES ($1);", c.StopList)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
