package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Command represents a command to execute, optionally with child commands
// that only run if the parent succeeds.
type Command struct {
	Cmd      string
	Children []Command
}

// UnmarshalYAML implements custom YAML unmarshaling for Command.
// It handles two forms:
//   - Scalar: "echo hello" → Command{Cmd: "echo hello"}
//   - Sequence: ["parent", "child1", "child2"] → Command with children
//     The first element is the parent command, the rest are children.
func (c *Command) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		c.Cmd = node.Value
		return nil

	case yaml.SequenceNode:
		if len(node.Content) == 0 {
			return fmt.Errorf("command sequence must not be empty")
		}

		// First element is the parent command (must be a scalar)
		first := node.Content[0]
		if first.Kind != yaml.ScalarNode {
			return fmt.Errorf("first element of command sequence must be a string")
		}
		c.Cmd = first.Value

		// Remaining elements are children
		for _, childNode := range node.Content[1:] {
			var child Command
			if err := child.UnmarshalYAML(childNode); err != nil {
				return err
			}
			c.Children = append(c.Children, child)
		}
		return nil

	default:
		return fmt.Errorf("unexpected YAML node kind %d for command", node.Kind)
	}
}

// PackageConfig holds the configuration for a single package.
type PackageConfig struct {
	Run map[string][]Command `yaml:"run"`
}

// Config represents the application configuration.
// It maps package names to their configuration.
type Config map[string]PackageConfig

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

// GetRun returns the directory-to-commands mapping for a given package name.
// Returns nil if the package is not configured.
func (c Config) GetRun(packageName string) map[string][]Command {
	pkg, ok := c[packageName]
	if !ok {
		return nil
	}
	return pkg.Run
}
