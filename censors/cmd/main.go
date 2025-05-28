// cmd/main.go
package main

import (
	"APIGateway/censors/config"
	"APIGateway/censors/pkg/api"
	"APIGateway/censors/pkg/middl"
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
	if err := godotenv.Load("censors/.env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	cfg := config.New()
	port := flag.String("censor-port", cfg.Censor.AdrPort, "Порт для censor сервиса")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	api := api.New()
	api.Router().Use(middl.Middle)

	server := &http.Server{
		Addr:    *port,
		Handler: api.Router(),
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Сервис цензуры запущен на %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	<-done
	log.Print("Сервис цензуры останавливается...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при остановке сервера: %v", err)
	}
	log.Print("Сервис цензуры остановлен")
}
