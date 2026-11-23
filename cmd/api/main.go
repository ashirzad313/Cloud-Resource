package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ashirzad313/Cloud-Resource/internal/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var pool *pgxpool.Pool
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var err error
		pool, err = pgxpool.New(ctx, databaseURL)
		if err == nil {
			err = pool.Ping(ctx)
		}
		cancel()

		if err != nil {
			if pool != nil {
				pool.Close()
			}
			log.Fatalf("connect to PostgreSQL: %v", err)
		}

		defer pool.Close()
		log.Println("connected to PostgreSQL")
	} else {
		log.Println("DATABASE_URL is not set; /ready will report unavailable")
	}

	// The HTTP layer only needs a Ping method, not the details of pgxpool.
	var database httpapi.DatabasePinger
	if pool != nil {
		database = pool
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(database),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
