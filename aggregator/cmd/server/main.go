package main

import (
	"APIGateway/aggregator/config"
	"APIGateway/aggregator/pkg/api"
	"APIGateway/aggregator/pkg/logger"
	"APIGateway/aggregator/pkg/middl"
	"APIGateway/aggregator/pkg/rss"
	"APIGateway/aggregator/pkg/storage"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func init() {
	if err := godotenv.Load("aggregator/.env"); err != nil {
		log.Print("No .env file found in aggregator/")
	}
}

func main() {
	// Инициализация логгера
	logInstance, err := logger.NewLogger("aggregator.log")
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	logInstance.InfoWithRequestID("Приложение запущено")

	// Загрузка конфигурации
	cfg := config.New()

	// Подключение к базе данных
	db, err := storage.ConnectByURL(cfg.URLdb)
	if err != nil {
		logInstance.ErrorWithRequestID("Ошибка подключения к БД: ", err)
		os.Exit(1)
	}
	defer db.Close()

	store := storage.NewStorage(db)

	// Загрузка конфигурации RSS
	rssConfig, err := storage.LoadConfig("aggregator/cmd/server/config.json", logInstance)
	if err != nil {
		logInstance.ErrorWithRequestID("Ошибка загрузки конфигурации RSS: ", err)
		os.Exit(1)
	}

	// Инициализация API и маршрутов
	apiHandler := api.NewAPI(store, logInstance)
	router := apiHandler.Router()

	// Подключение middleware
	router.Use(middl.Middle)
	router.Use(middl.WithRequestID)

	// Каналы для обмена сообщениями RSS
	postCh := make(chan storage.Post)
	errCh := make(chan error)

	// Запуск RSS-парсера
	go rss.StartPolling(rssConfig.RSS, rssConfig.RequestPeriod, postCh, errCh, logInstance)

	// Обработка полученных постов
	go func() {
		for post := range postCh {
			if err := store.SavePost(post, logInstance); err != nil {
				logInstance.ErrorWithRequestID("Ошибка сохранения поста:", err)
			}
		}
	}()

	// Обработка ошибок RSS
	go func() {
		for err := range errCh {
			logInstance.ErrorWithRequestID("Ошибка получения RSS: ", err)
		}
	}()

	logInstance.InfoWithRequestID("Сервер запущен на порту " + cfg.AdrPort)

	// Запуск HTTP-сервера
	if err := http.ListenAndServe(cfg.AdrPort, router); err != nil {
		logInstance.ErrorWithRequestID("Ошибка при запуске сервера: ", err)
		os.Exit(1)
	}
}
