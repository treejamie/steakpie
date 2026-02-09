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
  run:
    /opt/jamiec:
      - docker compose up
      - echo "test"

hello-world:
  run:
    /opt/hello:
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
	jamiecRun := cfg.GetRun("jamiec")
	if jamiecRun == nil {
		t.Fatal("expected jamiec to have run config")
	}
	commands := jamiecRun["/opt/jamiec"]
	if len(commands) != 2 {
		t.Errorf("expected 2 commands for jamiec, got %d", len(commands))
	}
	if commands[0].Cmd != "docker compose up" {
		t.Errorf("expected first command to be 'docker compose up', got '%s'", commands[0].Cmd)
	}

	// Check hello-world package
	helloRun := cfg.GetRun("hello-world")
	if helloRun == nil {
		t.Fatal("expected hello-world to have run config")
	}
	helloCommands := helloRun["/opt/hello"]
	if len(helloCommands) != 1 {
		t.Errorf("expected 1 command for hello-world, got %d", len(helloCommands))
	}
}

func TestLoad_NestedConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `mypackage:
  run:
    /opt/mypackage:
      - - docker compose pull
        - doppler run -- docker compose up -d
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	commands := cfg.GetRun("mypackage")["/opt/mypackage"]
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
  run:
    /opt/mixed:
      - - cmd1
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

	commands := cfg.GetRun("mixed")["/opt/mixed"]
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
  run:
    /opt/deep:
      - - level1
        - - level2
          - level3
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	commands := cfg.GetRun("deep")["/opt/deep"]
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

func TestLoad_EmptySequenceCommand(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `bad:
  run:
    /opt/bad:
      - []
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for empty sequence command, got nil")
	}

	if !strings.Contains(err.Error(), "command sequence must not be empty") {
		t.Errorf("error should mention empty sequence, got: %v", err)
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

func TestLoad_MultipleDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `mypackage:
  run:
    /opt/frontend:
      - npm run build
    /opt/backend:
      - go build ./...
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	run := cfg.GetRun("mypackage")
	if len(run) != 2 {
		t.Fatalf("expected 2 directories, got %d", len(run))
	}

	frontend := run["/opt/frontend"]
	if len(frontend) != 1 || frontend[0].Cmd != "npm run build" {
		t.Errorf("expected frontend command 'npm run build', got %v", frontend)
	}

	backend := run["/opt/backend"]
	if len(backend) != 1 || backend[0].Cmd != "go build ./..." {
		t.Errorf("expected backend command 'go build ./...', got %v", backend)
	}
}

func TestGetRun_ExistingPackage(t *testing.T) {
	cfg := Config{
		"test-package": {
			Run: map[string][]Command{
				"/opt/test": {
					{Cmd: "echo test"},
					{Cmd: "docker compose up"},
				},
			},
		},
		"other-package": {
			Run: map[string][]Command{
				"/opt/other": {
					{Cmd: "echo other"},
				},
			},
		},
	}

	run := cfg.GetRun("test-package")
	if run == nil {
		t.Fatal("expected run config for test-package")
	}
	commands := run["/opt/test"]
	if len(commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(commands))
	}
	if commands[0].Cmd != "echo test" {
		t.Errorf("expected first command to be 'echo test', got '%s'", commands[0].Cmd)
	}
}

func TestGetRun_NonExistingPackage(t *testing.T) {
	cfg := Config{
		"test-package": {
			Run: map[string][]Command{
				"/opt/test": {
					{Cmd: "echo test"},
				},
			},
		},
	}

	run := cfg.GetRun("nonexistent")
	if run != nil {
		t.Errorf("expected nil for non-existing package, got %v", run)
	}
}
