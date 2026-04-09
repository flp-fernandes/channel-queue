package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/flp-fernandes/product-views/internal/domain"
	"github.com/flp-fernandes/product-views/internal/httpapi"
	"github.com/flp-fernandes/product-views/internal/queue"
	"github.com/flp-fernandes/product-views/internal/repository"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load env file: %v", err)
	}

	dbCfg := repository.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := repository.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer db.Close()
	productViewsRepo := repository.NewProductViewsRepository(db)

	queueBufferSize, err := strconv.Atoi(os.Getenv("QUEUE_BUFFER"))
	if err != nil {
		log.Fatalf("failed to parse QUEUE_BUFFER: %v", err)
	}
	eventQueue := queue.NewEventQueue(queueBufferSize)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	numWorkers, err := strconv.Atoi(os.Getenv("WORKERS"))
	if err != nil {
		log.Fatalf("failed to parse WORKERS: %v", err)
	}
	eventQueue.StartWorkers(ctx, numWorkers, func(ctx context.Context, views []domain.ProductView) {
		if err := productViewsRepo.BulkInsert(ctx, views); err != nil {
			log.Printf("Failed to bulk insert %d product views: %v", len(views), err)
		}
	})

	handler := httpapi.NewHandler(eventQueue)
	router := httpapi.NewRouter(handler)

	server := &http.Server{
		Addr:         ":3000",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		log.Println("Server listening on :3000")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	cancel()

	log.Println("server stopped")

}
