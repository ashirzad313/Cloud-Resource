package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ashirzad313/Cloud-Resource/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
