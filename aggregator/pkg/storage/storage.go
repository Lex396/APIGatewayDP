package storage

import (
	"APIGateway/aggregator/pkg/logger"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Post struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Link    string    `json:"link"`
	Content string    `json:"content"`
	PubTime time.Time `json:"pub_time"`
}

type Storage struct {
	DB *sql.DB
}

type Pagination struct {
	NumOfPages int `json:"numOfPages,omitempty"`
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
}

type Config struct {
	RSS           []string `json:"rss"`
	RequestPeriod int      `json:"request_period"`
}

type StorageInterface interface {
	GetLastPosts(limit, page int) ([]Post, Pagination, error)
	SavePost(post Post, logInstance *logger.Logger) error
	GetPostsByTitle(search string, limit, offset int) ([]Post, Pagination, error)
	GetPostByID(id int) (Post, error)
	CountPostsByTitle(search string) (int, error)
}

// NewDatabase создает подключение к БД
func NewDatabase(host, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	return db, nil
}

// NewStorage создает новый экземпляр хранилища
func NewStorage(db *sql.DB) *Storage {
	return &Storage{DB: db}
}

// GetLastPosts возвращает последние n публикаций
func (s *Storage) GetLastPosts(limit, page int) ([]Post, Pagination, error) {
	var total int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM posts").Scan(&total)
	if err != nil {
		return nil, Pagination{}, fmt.Errorf("ошибка подсчета общего количества постов: %w", err)
	}

	if limit <= 0 {
		limit = 5
	}
	if page <= 0 {
		page = 1
	}

	totalPages := (total + limit - 1) / limit
	if page > totalPages && totalPages != 0 {
		page = totalPages
	}
	offset := (page - 1) * limit

	rows, err := s.DB.Query(
		`SELECT id, title, content, pub_time, link 
		 FROM posts 
		 ORDER BY pub_time DESC 
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, Pagination{}, fmt.Errorf("ошибка получения публикаций: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.PubTime, &post.Link); err != nil {
			return nil, Pagination{}, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		posts = append(posts, post)
	}

	pagination := Pagination{
		NumOfPages: totalPages,
		Page:       page,
		Limit:      limit,
	}

	return posts, pagination, nil
}

// SavePost сохраняет новость в БД
func (s *Storage) SavePost(post Post, logInstance *logger.Logger) error {
	_, err := s.DB.Exec(
		"INSERT INTO posts (title, content, pub_time, link) VALUES ($1, $2, $3, $4) ON CONFLICT (link) DO NOTHING",
		post.Title, post.Content, post.PubTime, post.Link,
	)
	if err != nil {
		logInstance.Error("Ошибка сохранения новости:", err)
		return err
	}
	return nil
}

// CountPostsByTitle получения количества новостей по фильтру
func (s *Storage) CountPostsByTitle(search string) (int, error) {
	query := "SELECT COUNT(*) FROM posts"
	var args []interface{}

	if search != "" {
		query += " WHERE title ILIKE $1"
		args = append(args, "%"+search+"%")
	}

	var count int
	err := s.DB.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка при подсчете количества постов: %w", err)
	}

	return count, nil
}

// GetPostsByTitle получение постов по фильтру с LIMIT и OFFSET, возвращает посты, пагинацию и ошибку.
func (s *Storage) GetPostsByTitle(search string, limit, offset int) ([]Post, Pagination, error) {
	query := "SELECT id, title, content, pub_time, link FROM posts"
	var args []interface{}
	argIndex := 1

	if search != "" {
		query += fmt.Sprintf(" WHERE title ILIKE $%d", argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY pub_time DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, Pagination{}, fmt.Errorf("ошибка получения постов: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.PubTime, &post.Link); err != nil {
			return nil, Pagination{}, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		posts = append(posts, post)
	}

	// Получаем общее количество постов по фильтру
	totalCount, err := s.CountPostsByTitle(search)
	if err != nil {
		return nil, Pagination{}, fmt.Errorf("ошибка подсчета постов: %w", err)
	}

	totalPages := (totalCount + limit - 1) / limit
	page := offset/limit + 1

	pagination := Pagination{
		NumOfPages: totalPages,
		Page:       page,
		Limit:      limit,
	}

	return posts, pagination, nil
}

// GetPostByID получение поста по id
func (s *Storage) GetPostByID(id int) (Post, error) {
	var post Post
	query := `SELECT id, title, link, content, pub_time FROM news WHERE id = $1`

	row := s.DB.QueryRow(query, id)
	err := row.Scan(&post.ID, &post.Title, &post.Link, &post.Content, &post.PubTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return post, fmt.Errorf("post with id %d not found", id)
		}
		return post, err
	}
	return post, nil
}

// LoadConfig загружает конфигурацию из JSON-файла
func LoadConfig(filename string, logInstance *logger.Logger) (*Config, error) {
	logInstance.Info("Загрузка конфигурации из файла:", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурационного файла: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфигурации: %w", err)
	}

	logInstance.Info("Загружена конфигурация:", config)
	return &config, nil
}
