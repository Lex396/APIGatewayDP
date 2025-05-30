// storage/postgres.go
package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, constr string) (*Store, error) {
	db, err := pgxpool.New(ctx, constr)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (p *Store) AddComment(c Comment) error {
	_, err := p.db.Exec(context.Background(), `
    INSERT INTO comments (news_id, parent_id, content) 
    VALUES ($1, $2, $3)`, c.NewsID, c.ParentID, c.Content)
	if err != nil {
		fmt.Println("[AddComment ERROR]", err)
	}
	return err
}

func (p *Store) AllComments(newsID int) ([]Comment, error) {
	rows, err := p.db.Query(context.Background(), `
    SELECT id, news_id, parent_id, content, pubtime 
    FROM comments 
    WHERE news_id = $1
    ORDER BY pubtime DESC`, newsID)
	if err != nil {
		fmt.Println("[AllComments ERROR]", err)
		return nil, err
	}

	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		err = rows.Scan(&c.ID, &c.NewsID, &c.ParentID, &c.Content, &c.PubTime)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (p *Store) DeleteComment(id int) error {
	_, err := p.db.Exec(context.Background(), `DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		fmt.Println("[DeleteComment ERROR]", err)
	}
	return err
}
