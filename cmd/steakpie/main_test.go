package main

import (
	"os"
	"path/filepath"
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

func TestRun_MissingConfigArgument(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set args to only contain program name
	os.Args = []string{"steakpie"}

	// Run should return an error
	err := run()
	if err == nil {
		t.Fatal("expected error when config file argument is missing, got nil")
	}

	// Error message should mention config file is required
	errMsg := err.Error()
	if !strings.Contains(errMsg, "config file path is required") {
		t.Errorf("error message should mention config file requirement, got: %s", errMsg)
	}

	// Error message should provide usage information
	if !strings.Contains(errMsg, "Usage:") {
		t.Errorf("error message should include usage information, got: %s", errMsg)
	}
}

func TestRun_InvalidConfigPath(t *testing.T) {
	// Save original args and env var, restore after test
	originalArgs := os.Args
	originalSecret := os.Getenv("WEBHOOK_SECRET")
	defer func() {
		os.Args = originalArgs
		if originalSecret != "" {
			os.Setenv("WEBHOOK_SECRET", originalSecret)
		} else {
			os.Unsetenv("WEBHOOK_SECRET")
		}
	}()

	// Set WEBHOOK_SECRET
	os.Setenv("WEBHOOK_SECRET", "test-secret")

	// Set args with non-existent config file
	os.Args = []string{"steakpie", "/nonexistent/config.yaml"}

	// Run should return an error
	err := run()
	if err == nil {
		t.Fatal("expected error when config file doesn't exist, got nil")
	}

	// Error message should mention config loading failure
	errMsg := err.Error()
	if !strings.Contains(errMsg, "failed to load config") {
		t.Errorf("error message should mention config loading failure, got: %s", errMsg)
	}
}

func TestRun_InvalidYAMLConfig(t *testing.T) {
	// Save original args and env var, restore after test
	originalArgs := os.Args
	originalSecret := os.Getenv("WEBHOOK_SECRET")
	defer func() {
		os.Args = originalArgs
		if originalSecret != "" {
			os.Setenv("WEBHOOK_SECRET", originalSecret)
		} else {
			os.Unsetenv("WEBHOOK_SECRET")
		}
	}()

	// Set WEBHOOK_SECRET
	os.Setenv("WEBHOOK_SECRET", "test-secret")

	// Create a temporary invalid YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	invalidYAML := "invalid: yaml: content:\n  - [unclosed"
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to create test config file: %v", err)
	}

	// Set args with invalid config file
	os.Args = []string{"steakpie", configPath}

	// Run should return an error
	err := run()
	if err == nil {
		t.Fatal("expected error when config file is invalid YAML, got nil")
	}

	// Error message should mention config loading failure
	errMsg := err.Error()
	if !strings.Contains(errMsg, "failed to load config") {
		t.Errorf("error message should mention config loading failure, got: %s", errMsg)
	}
}
