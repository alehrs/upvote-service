package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"upvote-service/internal/db"
	"upvote-service/internal/handlers"
	"upvote-service/internal/repository"
	"upvote-service/internal/server"
	"upvote-service/internal/service"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("load .env: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	repo := repository.New(pool)
	svc := service.New(repo)
	handler := handlers.New(svc)
	srv := server.New(handler, addr)

	log.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
