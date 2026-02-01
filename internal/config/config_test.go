package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidYAML(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `jamiec:
  - docker compose up
  - echo "test"

hello-world:
  - echo "hello"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(cfg) != 2 {
		t.Errorf("expected 2 packages, got %d", len(cfg))
	}

	// Check jamiec package
	jamiecCommands := cfg["jamiec"]
	if len(jamiecCommands) != 2 {
		t.Errorf("expected 2 commands for jamiec, got %d", len(jamiecCommands))
	}
	if jamiecCommands[0] != "docker compose up" {
		t.Errorf("expected first command to be 'docker compose up', got '%s'", jamiecCommands[0])
	}

	// Check hello-world package
	helloCommands := cfg["hello-world"]
	if len(helloCommands) != 1 {
		t.Errorf("expected 1 command for hello-world, got %d", len(helloCommands))
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `invalid yaml content
	this is not valid:
	  - [unclosed bracket
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse config file") {
		t.Errorf("error should mention parse failure, got: %v", err)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("error should mention read failure, got: %v", err)
	}
}

func TestLoad_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write empty file
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for empty config, got nil")
	}

	if !strings.Contains(err.Error(), "config file is empty") {
		t.Errorf("error should mention empty config, got: %v", err)
	}
}

func TestGetCommands_ExistingPackage(t *testing.T) {
	cfg := Config{
		"test-package": []string{"echo test", "docker compose up"},
		"other-package": []string{"echo other"},
	}

	commands := cfg.GetCommands("test-package")
	if len(commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(commands))
	}
	if commands[0] != "echo test" {
		t.Errorf("expected first command to be 'echo test', got '%s'", commands[0])
	}
}

func TestGetCommands_NonExistingPackage(t *testing.T) {
	cfg := Config{
		"test-package": []string{"echo test"},
	}

	commands := cfg.GetCommands("nonexistent")
	if commands != nil {
		t.Errorf("expected nil for non-existing package, got %v", commands)
	}
}
