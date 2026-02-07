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
//   - Scalar: "docker compose pull" → Command{Cmd: "docker compose pull"}
//   - Mapping: "docker compose pull": ["child1", "child2"] → Command with children
func (c *Command) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		c.Cmd = node.Value
		return nil

	case yaml.MappingNode:
		if len(node.Content) != 2 {
			return fmt.Errorf("command mapping must have exactly one key, got %d", len(node.Content)/2)
		}

		// First element is the key (command string)
		keyNode := node.Content[0]
		if keyNode.Kind != yaml.ScalarNode {
			return fmt.Errorf("command key must be a string")
		}
		c.Cmd = keyNode.Value

		// Second element is the value (list of child commands)
		valueNode := node.Content[1]
		if valueNode.Kind != yaml.SequenceNode {
			return fmt.Errorf("command children must be a list")
		}

		var children []Command
		for _, childNode := range valueNode.Content {
			var child Command
			if err := child.UnmarshalYAML(childNode); err != nil {
				return err
			}
			children = append(children, child)
		}
		c.Children = children
		return nil

	default:
		return fmt.Errorf("unexpected YAML node kind %d for command", node.Kind)
	}
}

// Config represents the application configuration.
// It maps package names to a list of commands to execute.
type Config map[string][]Command

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
func (c Config) GetCommands(packageName string) []Command {
	return c[packageName]
}
