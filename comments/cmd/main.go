// main.go
package main

import (
	"APIGateway/aggregator/pkg/logger"
	"APIGateway/comments/config"
	"APIGateway/comments/pkg/api"
	"APIGateway/comments/pkg/middl"
	"APIGateway/comments/pkg/storage"
	"context"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	if err := godotenv.Load("comments/.env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	logg, err := logger.NewLogger("comments.log")
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err) // Логируем через стандартный, т.к. логгер ещё не инициализирован
	}
	logg.InfoWithRequestID("Comments сервис запущен")

	cfg := config.New()
	port := flag.String("comments-port", cfg.AdrPort, "Порт для comments сервиса")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := storage.New(ctx, cfg.URLdb)
	if err != nil {
		logg.ErrorWithRequestID("Ошибка подключения к БД:", err)
		os.Exit(1)
	}

	api := api.New(db, logg)
	api.Router().Use(middl.Middle)

	server := &http.Server{
		Addr:    *port,
		Handler: api.Router(),
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logg.InfoWithRequestID("Сервер комментариев запущен на " + *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.ErrorWithRequestID("Ошибка сервера:", err)
		}
	}()

	<-done
	logg.InfoWithRequestID("Сервер комментариев останавливается...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logg.ErrorWithRequestID("Ошибка при остановке сервера:", err)
	}
	logg.InfoWithRequestID("Сервер комментариев остановлен")
}
