package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	httpapi "dota2-hero-grid-maker-back/internal/http"
	"dota2-hero-grid-maker-back/internal/storage"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	db, err := storage.Open(context.Background(), storage.DBConfig{DSN: dsn})
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}

	h := httpapi.NewRouter(&httpapi.Handler{
		Grids:    storage.NewGridRepository(db),
		Sessions: storage.NewSessionStore(db),
		Users:    storage.NewUserStore(db),
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
