package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ashirzad313/Cloud-Resource/internal/httpapi"
	"github.com/ashirzad313/Cloud-Resource/internal/storage/postgres"
	"github.com/ashirzad313/Cloud-Resource/internal/vm"
	"github.com/ashirzad313/Cloud-Resource/migrations"
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

		migrationCtx, cancelMigrations := context.WithTimeout(context.Background(), 30*time.Second)
		applied, err := migrations.Apply(migrationCtx, pool)
		cancelMigrations()
		if err != nil {
			pool.Close()
			log.Fatalf("database migrations failed: %v", err)
		}

		if len(applied) == 0 {
			log.Println("database migrations up to date")
		} else {
			for _, name := range applied {
				log.Printf("applied migration: %s", name)
			}
		}
	} else {
		log.Println("DATABASE_URL is not set; /ready will report unavailable")
	}

	// Assemble application dependencies in one place.
	var database httpapi.DatabasePinger
	var vmService *vm.Service
	if pool != nil {
		database = pool
		repository := postgres.NewVMRepository(pool)
		vmService = vm.NewService(repository)
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(database, vmService),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
