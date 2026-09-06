package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/busrau/calc_services/internal/calculator"
	"github.com/busrau/calc_services/internal/httpapi"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewMux(calculator.New()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("calculator service listening on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
