package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"clima-por-cep/internal/weather"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           weather.NewHandler(weather.NewService(nil)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("server listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
