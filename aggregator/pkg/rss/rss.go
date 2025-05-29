package rss

import (
	"APIGateway/aggregator/pkg/logger"
	"APIGateway/aggregator/pkg/storage"
	"encoding/xml"
	"io"
	"net/http"
	"time"
)

type RSS struct {
	Channel struct {
		Items []Item `xml:"item"`
	} `xml:"channel"`
}

type Item struct {
	Title   string `xml:"title"`
	Content string `xml:"description"`
	PubTime string `xml:"pubDate"`
	Link    string `xml:"link"`
}

var timeFormats = []string{
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon 2 Jan 2006 15:04:05 GMT",
	"Mon, 2 Jan 2006 15:04:05 MST",
	"Mon 2 Jan 2006 15:04:05 MST",
}

// parseRSS загружает и парсит RSS-ленту
func parseRSS(url string, logInstance *logger.Logger) ([]storage.Post, error) {
	logInstance.InfoWithRequestID("Запрос к RSS-ленте:", url)

	resp, err := http.Get(url)
	if err != nil {
		logInstance.ErrorWithRequestID("Ошибка HTTP-запроса:", err)
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logInstance.ErrorWithRequestID("Ошибка чтения тела ответа:", err)
		return nil, err
	}

	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		logInstance.ErrorWithRequestID("Ошибка парсинга XML:", err)
		return nil, err
	}

	var posts []storage.Post
	for _, item := range rss.Channel.Items {
		pubTime := parsePubDate(item.PubTime, logInstance)

		posts = append(posts, storage.Post{
			Title:   item.Title,
			Content: item.Content,
			PubTime: pubTime,
			Link:    item.Link,
		})
	}

	return posts, nil
}

// parsePubDate пытается разобрать дату в разных форматах
func parsePubDate(dateStr string, logInstance *logger.Logger) time.Time {
	for _, format := range timeFormats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t
		}
	}
	logInstance.ErrorWithRequestID("Ошибка парсинга даты:", dateStr)
	return time.Now()
}

// StartPolling запускает горутины для чтения RSS с использованием каналов
func StartPolling(urls []string, period int, postChan chan<- storage.Post, errChan chan<- error, logInstance *logger.Logger) {
	for _, url := range urls {
		go func(feed string) {
			pollingRSS(feed, postChan, errChan, logInstance)
			ticker := time.NewTicker(time.Duration(period) * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				pollingRSS(feed, postChan, errChan, logInstance)
			}
		}(url)
	}
}

// pollingRSS выполняет запрос к RSS-ленте и отправляет результаты в каналы
func pollingRSS(feed string, postChan chan<- storage.Post, errChan chan<- error, logInstance *logger.Logger) {
	logInstance.InfoWithRequestID("Читаем RSS:", feed)
	posts, err := parseRSS(feed, logInstance)
	if err != nil {
		errChan <- err
		return
	}

	if len(posts) == 0 {
		logInstance.InfoWithRequestID("Нет новых статей из:", feed)
	}

	for _, post := range posts {
		postChan <- post
	}
}
