package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jc/steakpie/internal/webhook"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("WEBHOOK_SECRET environment variable is required")
	}

	http.Handle("/version/1", webhook.Handler([]byte(secret)))

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
