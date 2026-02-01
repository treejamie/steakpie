package main

import (
	"os"
	"strings"
	"testing"
)

func TestRun_MissingWebhookSecret(t *testing.T) {
	// Save original env var and restore after test
	originalSecret := os.Getenv("WEBHOOK_SECRET")
	defer func() {
		if originalSecret != "" {
			os.Setenv("WEBHOOK_SECRET", originalSecret)
		}
	}()

	// Ensure WEBHOOK_SECRET is not set
	os.Unsetenv("WEBHOOK_SECRET")

	// Run should return an error
	err := run()
	if err == nil {
		t.Fatal("expected error when WEBHOOK_SECRET is not set, got nil")
	}

	// Error message should mention the missing variable
	errMsg := err.Error()
	if !strings.Contains(errMsg, "WEBHOOK_SECRET") {
		t.Errorf("error message should mention WEBHOOK_SECRET, got: %s", errMsg)
	}

	// Error message should provide helpful guidance
	if !strings.Contains(errMsg, "environment variable is required") {
		t.Errorf("error message should explain it's required, got: %s", errMsg)
	}
}

func TestRun_EmptyWebhookSecret(t *testing.T) {
	// Save original env var and restore after test
	originalSecret := os.Getenv("WEBHOOK_SECRET")
	defer func() {
		if originalSecret != "" {
			os.Setenv("WEBHOOK_SECRET", originalSecret)
		} else {
			os.Unsetenv("WEBHOOK_SECRET")
		}
	}()

	// Set WEBHOOK_SECRET to empty string
	os.Setenv("WEBHOOK_SECRET", "")

	// Run should return an error
	err := run()
	if err == nil {
		t.Fatal("expected error when WEBHOOK_SECRET is empty, got nil")
	}

	// Error message should mention the missing variable
	errMsg := err.Error()
	if !strings.Contains(errMsg, "WEBHOOK_SECRET") {
		t.Errorf("error message should mention WEBHOOK_SECRET, got: %s", errMsg)
	}
}
