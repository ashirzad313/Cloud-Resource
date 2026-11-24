// vmcheck is a temporary learning tool to exercise the repository without
// creating REST endpoints yet. It will be removed when the VM API is available.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ashirzad313/Cloud-Resource/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("configure PostgreSQL pool: %v", err)
	}
	defer pool.Close()

	machines, err := postgres.NewVMRepository(pool).List(ctx)
	if err != nil {
		log.Fatalf("list virtual machines: %v", err)
	}

	fmt.Printf("Found %d virtual machine(s)\n", len(machines))
	for _, machine := range machines {
		fmt.Printf("%s | %s | %d CPU | %d MB | %s | %s\n",
			machine.ID, machine.Name, machine.CPUCores,
			machine.MemoryMB, machine.Region, machine.Status,
		)
	}
}
