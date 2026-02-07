package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// chdir changes to the given directory and returns a cleanup function
// that restores the original working directory.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

// setEnv sets an environment variable for the duration of the test.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	orig, ok := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if ok {
			os.Setenv(key, orig)
		} else {
			os.Unsetenv(key)
		}
	})
}

// unsetEnv unsets an environment variable for the duration of the test.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	orig, ok := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if ok {
			os.Setenv(key, orig)
		}
	})
}

// writeFile creates a file with the given content in dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

const validConfig = "mypackage:\n  - echo hello\n"

func TestRun_MissingWebhookSecret(t *testing.T) {
	unsetEnv(t, "WEBHOOK_SECRET")

	err := run()
	if err == nil {
		t.Fatal("expected error when WEBHOOK_SECRET is not set, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "WEBHOOK_SECRET") {
		t.Errorf("error message should mention WEBHOOK_SECRET, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "environment variable is required") {
		t.Errorf("error message should explain it's required, got: %s", errMsg)
	}
}

func TestRun_EmptyWebhookSecret(t *testing.T) {
	setEnv(t, "WEBHOOK_SECRET", "")

	err := run()
	if err == nil {
		t.Fatal("expected error when WEBHOOK_SECRET is empty, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "WEBHOOK_SECRET") {
		t.Errorf("error message should mention WEBHOOK_SECRET, got: %s", errMsg)
	}
}

func TestRun_NoConfigFilePresent(t *testing.T) {
	setEnv(t, "WEBHOOK_SECRET", "test-secret")
	chdir(t, t.TempDir())

	err := run()
	if err == nil {
		t.Fatal("expected error when no config file exists, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "no config file found") {
		t.Errorf("error message should say no config file found, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "config.yml") {
		t.Errorf("error message should mention config.yml, got: %s", errMsg)
	}
}

func TestFindConfig_YmlFound(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "config.yml", validConfig)
	chdir(t, dir)

	path, err := findConfig()
	if err != nil {
		t.Fatal("expected config.yml to be found, got error:", err)
	}
	if path != "config.yml" {
		t.Errorf("expected path config.yml, got %s", path)
	}
}

func TestFindConfig_YamlFound(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "config.yaml", validConfig)
	chdir(t, dir)

	path, err := findConfig()
	if err != nil {
		t.Fatal("expected config.yaml to be found, got error:", err)
	}
	if path != "config.yaml" {
		t.Errorf("expected path config.yaml, got %s", path)
	}
}

func TestFindConfig_YmlTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "config.yml", validConfig)
	writeFile(t, dir, "config.yaml", validConfig)
	chdir(t, dir)

	path, err := findConfig()
	if err != nil {
		t.Fatal("expected config.yml to be found, got error:", err)
	}
	if path != "config.yml" {
		t.Errorf("expected config.yml to take precedence, got %s", path)
	}
}

func TestRun_InvalidConfigPath(t *testing.T) {
	dir := t.TempDir()
	// Create a config.yml that is completely empty (will fail validation)
	writeFile(t, dir, "config.yml", "")

	setEnv(t, "WEBHOOK_SECRET", "test-secret")
	chdir(t, dir)

	err := run()
	if err == nil {
		t.Fatal("expected error when config file is empty, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "failed to load config") {
		t.Errorf("error message should mention config loading failure, got: %s", errMsg)
	}
}

func TestRun_InvalidYAMLConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "config.yml", "invalid: yaml: content:\n  - [unclosed")

	setEnv(t, "WEBHOOK_SECRET", "test-secret")
	chdir(t, dir)

	err := run()
	if err == nil {
		t.Fatal("expected error when config file is invalid YAML, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "failed to load config") {
		t.Errorf("error message should mention config loading failure, got: %s", errMsg)
	}
}
