package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidYAML(t *testing.T) {
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
	if jamiecCommands[0].Cmd != "docker compose up" {
		t.Errorf("expected first command to be 'docker compose up', got '%s'", jamiecCommands[0].Cmd)
	}

	// Check hello-world package
	helloCommands := cfg["hello-world"]
	if len(helloCommands) != 1 {
		t.Errorf("expected 1 command for hello-world, got %d", len(helloCommands))
	}
}

func TestLoad_NestedConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `mypackage:
  - docker compose pull:
      - doppler run -- docker compose up -d
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	commands := cfg["mypackage"]
	if len(commands) != 1 {
		t.Fatalf("expected 1 top-level command, got %d", len(commands))
	}

	parent := commands[0]
	if parent.Cmd != "docker compose pull" {
		t.Errorf("expected parent cmd 'docker compose pull', got '%s'", parent.Cmd)
	}
	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child command, got %d", len(parent.Children))
	}
	if parent.Children[0].Cmd != "doppler run -- docker compose up -d" {
		t.Errorf("expected child cmd 'doppler run -- docker compose up -d', got '%s'", parent.Children[0].Cmd)
	}
}

func TestLoad_MixedConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `mixed:
  - cmd1:
      - cmd2
  - cmd3
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	commands := cfg["mixed"]
	if len(commands) != 2 {
		t.Fatalf("expected 2 top-level commands, got %d", len(commands))
	}

	// First command has a child
	if commands[0].Cmd != "cmd1" {
		t.Errorf("expected first cmd 'cmd1', got '%s'", commands[0].Cmd)
	}
	if len(commands[0].Children) != 1 {
		t.Fatalf("expected 1 child for cmd1, got %d", len(commands[0].Children))
	}
	if commands[0].Children[0].Cmd != "cmd2" {
		t.Errorf("expected child cmd 'cmd2', got '%s'", commands[0].Children[0].Cmd)
	}

	// Second command is flat
	if commands[1].Cmd != "cmd3" {
		t.Errorf("expected second cmd 'cmd3', got '%s'", commands[1].Cmd)
	}
	if len(commands[1].Children) != 0 {
		t.Errorf("expected no children for cmd3, got %d", len(commands[1].Children))
	}
}

func TestLoad_DeeplyNestedConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `deep:
  - level1:
      - level2:
          - level3
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	commands := cfg["deep"]
	if len(commands) != 1 {
		t.Fatalf("expected 1 top-level command, got %d", len(commands))
	}

	l1 := commands[0]
	if l1.Cmd != "level1" {
		t.Errorf("expected 'level1', got '%s'", l1.Cmd)
	}
	if len(l1.Children) != 1 {
		t.Fatalf("expected 1 child at level1, got %d", len(l1.Children))
	}

	l2 := l1.Children[0]
	if l2.Cmd != "level2" {
		t.Errorf("expected 'level2', got '%s'", l2.Cmd)
	}
	if len(l2.Children) != 1 {
		t.Fatalf("expected 1 child at level2, got %d", len(l2.Children))
	}

	l3 := l2.Children[0]
	if l3.Cmd != "level3" {
		t.Errorf("expected 'level3', got '%s'", l3.Cmd)
	}
	if len(l3.Children) != 0 {
		t.Errorf("expected no children at level3, got %d", len(l3.Children))
	}
}

func TestLoad_InvalidNesting_MultipleKeys(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `bad:
  - cmd1: [child1]
    cmd2: [child2]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for multiple keys in mapping, got nil")
	}

	if !strings.Contains(err.Error(), "exactly one key") {
		t.Errorf("error should mention 'exactly one key', got: %v", err)
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
		"test-package": {
			{Cmd: "echo test"},
			{Cmd: "docker compose up"},
		},
		"other-package": {
			{Cmd: "echo other"},
		},
	}

	commands := cfg.GetCommands("test-package")
	if len(commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(commands))
	}
	if commands[0].Cmd != "echo test" {
		t.Errorf("expected first command to be 'echo test', got '%s'", commands[0].Cmd)
	}
}

func TestGetCommands_NonExistingPackage(t *testing.T) {
	cfg := Config{
		"test-package": {
			{Cmd: "echo test"},
		},
	}

	commands := cfg.GetCommands("nonexistent")
	if commands != nil {
		t.Errorf("expected nil for non-existing package, got %v", commands)
	}
}
