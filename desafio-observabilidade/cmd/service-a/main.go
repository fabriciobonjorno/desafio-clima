package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"clima-por-cep/internal/input"
	"clima-por-cep/internal/server"
	"clima-por-cep/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	serviceBURL := os.Getenv("SERVICE_B_URL")
	if serviceBURL == "" {
		serviceBURL = "http://service-b:8081/"
	}

	client := &http.Client{
		Timeout:   20 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	handler := telemetry.HTTPHandler("service-a", input.NewHandler(serviceBURL, client))
	if err := server.Run("service-a", "8080", handler); err != nil {
		log.Fatal(err)
	}
}
