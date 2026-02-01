package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
// It maps package names to a list of commands to execute.
type Config map[string][]string

// Load reads and parses a YAML configuration file.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(cfg) == 0 {
		return nil, fmt.Errorf("config file is empty")
	}

	return cfg, nil
}

// GetCommands returns the list of commands for a given package name.
// Returns nil if the package is not configured.
func (c Config) GetCommands(packageName string) []string {
	return c[packageName]
}
