package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jc/steakpie/internal/webhook"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n\n", err)
		os.Exit(1)
	}
}

func run() error {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		return fmt.Errorf("WEBHOOK_SECRET environment variable is required\n\n" +
			"Please set it before running:\n" +
			"  WEBHOOK_SECRET=your-secret-here ./steakpie\n\n" +
			"This secret is used to verify webhook signatures from GitHub.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.Handle("/version/1", webhook.Handler([]byte(secret)))

	log.Printf("✓ Server starting on port %s", port)
	log.Printf("✓ Webhook endpoint: http://localhost:%s/version/1", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
