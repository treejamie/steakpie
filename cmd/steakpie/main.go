package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jc/steakpie/internal/config"
	"github.com/jc/steakpie/internal/executor"
	"github.com/jc/steakpie/internal/webhook"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n\n", err)
		os.Exit(1)
	}
}

func findConfig() (string, error) {
	for _, name := range []string{"config.yml", "config.yaml"} {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("no config file found\n\n" +
		"Place a config.yml (or config.yaml) in the current directory.\n\n" +
		"Example:\n" +
		"  WEBHOOK_SECRET=secret steakpie\n\n" +
		"Optional environment variables:\n" +
		"  DB_PATH - Path to SQLite database (default: db.sqlite)")
}

func run() error {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		return fmt.Errorf("WEBHOOK_SECRET environment variable is required\n\n" +
			"Please set it before running:\n" +
			"  WEBHOOK_SECRET=your-secret-here steakpie\n\n" +
			"This secret is used to verify webhook signatures from GitHub.")
	}

	configPath, err := findConfig()
	if err != nil {
		return err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log.Printf("✓ Loaded config with %d package(s)", len(cfg))

	// Initialize event store for webhook deduplication
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "db.sqlite"
	}

	store, err := webhook.NewEventStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize event store: %w", err)
	}
	defer store.Close()

	log.Printf("✓ Initialized event store at %s", dbPath)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	runner := executor.ShellRunner{}
	http.Handle("/version/1", webhook.Handler([]byte(secret), cfg, store, runner))

	log.Printf("✓ Server starting on port %s", port)
	log.Printf("✓ Webhook endpoint: http://localhost:%s/version/1", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
