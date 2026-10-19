package main

import (
	"log"

	"clima-por-cep/internal/server"
	"clima-por-cep/internal/telemetry"
	"clima-por-cep/internal/weather"
)

func main() {
	handler := telemetry.HTTPHandler("service-b", weather.NewHandler(weather.NewService(nil)))
	if err := server.Run("service-b", "8081", handler); err != nil {
		log.Fatal(err)
	}
}
