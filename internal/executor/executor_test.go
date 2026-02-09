package executor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jc/steakpie/internal/config"
)

// MockRunner records commands and returns preset results.
type MockRunner struct {
	Commands []string
	Dirs     []string
	Results  map[string]struct {
		Output string
		Err    error
	}
}

func NewMockRunner() *MockRunner {
	return &MockRunner{
		Results: make(map[string]struct {
			Output string
			Err    error
		}),
	}
}

func (m *MockRunner) SetResult(cmd, output string, err error) {
	m.Results[cmd] = struct {
		Output string
		Err    error
	}{output, err}
}

func (m *MockRunner) Run(cmd string, dir string) (string, error) {
	m.Commands = append(m.Commands, cmd)
	m.Dirs = append(m.Dirs, dir)
	if r, ok := m.Results[cmd]; ok {
		return r.Output, r.Err
	}
	return "", nil
}

// captureLog captures log output during f() execution.
func captureLog(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(nil)
		log.SetFlags(log.LstdFlags)
	}()
	f()
	return buf.String()
}

// helper to wrap commands in a single-directory map
func dirCommands(dir string, commands []config.Command) map[string][]config.Command {
	return map[string][]config.Command{dir: commands}
}

func TestParallelFlow_FirstFails_SecondStillRuns(t *testing.T) {
	runner := NewMockRunner()
	runner.SetResult("cmd1", "", fmt.Errorf("exit status 1"))
	runner.SetResult("cmd2", "", nil)

	commands := dirCommands("/opt/test", []config.Command{
		{Cmd: "cmd1"},
		{Cmd: "cmd2"},
	})

	Execute(runner, "test-pkg", "delivery-1", commands)

	if len(runner.Commands) != 2 {
		t.Fatalf("expected 2 commands to run, got %d", len(runner.Commands))
	}
	if runner.Commands[0] != "cmd1" {
		t.Errorf("expected first command 'cmd1', got '%s'", runner.Commands[0])
	}
	if runner.Commands[1] != "cmd2" {
		t.Errorf("expected second command 'cmd2', got '%s'", runner.Commands[1])
	}
}

func TestNestedFlow_ParentFails_ChildSkipped(t *testing.T) {
	runner := NewMockRunner()
	runner.SetResult("parent", "", fmt.Errorf("exit status 1"))

	commands := dirCommands("/opt/test", []config.Command{
		{Cmd: "parent", Children: []config.Command{
			{Cmd: "child"},
		}},
	})

	Execute(runner, "test-pkg", "delivery-2", commands)

	if len(runner.Commands) != 1 {
		t.Fatalf("expected 1 command to run (child skipped), got %d: %v", len(runner.Commands), runner.Commands)
	}
	if runner.Commands[0] != "parent" {
		t.Errorf("expected 'parent', got '%s'", runner.Commands[0])
	}
}

func TestNestedFlow_ParentSucceeds_ChildRuns(t *testing.T) {
	runner := NewMockRunner()

	commands := dirCommands("/opt/test", []config.Command{
		{Cmd: "parent", Children: []config.Command{
			{Cmd: "child"},
		}},
	})

	Execute(runner, "test-pkg", "delivery-3", commands)

	if len(runner.Commands) != 2 {
		t.Fatalf("expected 2 commands to run, got %d", len(runner.Commands))
	}
	if runner.Commands[0] != "parent" {
		t.Errorf("expected first 'parent', got '%s'", runner.Commands[0])
	}
	if runner.Commands[1] != "child" {
		t.Errorf("expected second 'child', got '%s'", runner.Commands[1])
	}
}

func TestMixedFlow_ParentFails_ChildSkipped_SiblingRuns(t *testing.T) {
	runner := NewMockRunner()
	runner.SetResult("parent", "", fmt.Errorf("exit status 1"))

	commands := dirCommands("/opt/test", []config.Command{
		{Cmd: "parent", Children: []config.Command{
			{Cmd: "child"},
		}},
		{Cmd: "sibling"},
	})

	Execute(runner, "test-pkg", "delivery-4", commands)

	if len(runner.Commands) != 2 {
		t.Fatalf("expected 2 commands (parent+sibling, child skipped), got %d: %v", len(runner.Commands), runner.Commands)
	}
	if runner.Commands[0] != "parent" {
		t.Errorf("expected first 'parent', got '%s'", runner.Commands[0])
	}
	if runner.Commands[1] != "sibling" {
		t.Errorf("expected second 'sibling', got '%s'", runner.Commands[1])
	}
}

func TestDeeplyNestedChain(t *testing.T) {
	runner := NewMockRunner()

	commands := dirCommands("/opt/test", []config.Command{
		{Cmd: "l1", Children: []config.Command{
			{Cmd: "l2", Children: []config.Command{
				{Cmd: "l3"},
			}},
		}},
	})

	Execute(runner, "test-pkg", "delivery-5", commands)

	expected := []string{"l1", "l2", "l3"}
	if len(runner.Commands) != len(expected) {
		t.Fatalf("expected %d commands, got %d: %v", len(expected), len(runner.Commands), runner.Commands)
	}
	for i, exp := range expected {
		if runner.Commands[i] != exp {
			t.Errorf("command %d: expected '%s', got '%s'", i, exp, runner.Commands[i])
		}
	}
}

func TestEmptyCommandList(t *testing.T) {
	runner := NewMockRunner()

	Execute(runner, "test-pkg", "delivery-6", map[string][]config.Command{})

	if len(runner.Commands) != 0 {
		t.Errorf("expected no commands to run, got %d", len(runner.Commands))
	}
}

func TestDirectoryPassedToRunner(t *testing.T) {
	runner := NewMockRunner()

	commands := map[string][]config.Command{
		"/opt/myapp": {
			{Cmd: "echo hello"},
		},
	}

	Execute(runner, "test-pkg", "delivery-7", commands)

	if len(runner.Dirs) != 1 {
		t.Fatalf("expected 1 dir, got %d", len(runner.Dirs))
	}
	if runner.Dirs[0] != "/opt/myapp" {
		t.Errorf("expected dir '/opt/myapp', got '%s'", runner.Dirs[0])
	}
}

func TestLogOutput_ParallelFlow(t *testing.T) {
	runner := NewMockRunner()
	runner.SetResult("cmd1", "hello output", nil)

	commands := dirCommands("/opt/test", []config.Command{
		{Cmd: "cmd1"},
		{Cmd: "cmd2"},
	})

	output := captureLog(func() {
		Execute(runner, "mypkg", "d-123", commands)
	})

	expectations := []string{
		"start webhook for mypkg received with id: d-123",
		"executing in directory: /opt/test",
		"running command 1 of 2: cmd1",
		"output: hello output",
		"command 1 of 2 succeeded",
		"running command 2 of 2: cmd2",
		"command 2 of 2 succeeded",
		"end webhook for mypkg with id: d-123",
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected log to contain %q, got:\n%s", exp, output)
		}
	}
}

func TestLogOutput_NestedFlow(t *testing.T) {
	runner := NewMockRunner()

	commands := dirCommands("/opt/test", []config.Command{
		{Cmd: "parent", Children: []config.Command{
			{Cmd: "child"},
		}},
	})

	output := captureLog(func() {
		Execute(runner, "mypkg", "d-456", commands)
	})

	if !strings.Contains(output, "executing in directory: /opt/test") {
		t.Errorf("expected directory log, got:\n%s", output)
	}
}

func TestLogOutput_FailedCommand(t *testing.T) {
	runner := NewMockRunner()
	runner.SetResult("bad", "", fmt.Errorf("exit status 1"))

	commands := dirCommands("/opt/test", []config.Command{
		{Cmd: "bad"},
	})

	output := captureLog(func() {
		Execute(runner, "mypkg", "d-789", commands)
	})

	if !strings.Contains(output, "command 1 of 1 failed") {
		t.Errorf("expected failure log, got:\n%s", output)
	}
}

func TestShellRunner_Integration(t *testing.T) {
	runner := ShellRunner{}

	output, err := runner.Run("echo hello world", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	trimmed := strings.TrimSpace(output)
	if trimmed != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", trimmed)
	}
}

func TestShellRunner_Integration_FailingCommand(t *testing.T) {
	runner := ShellRunner{}

	_, err := runner.Run("false", "")
	if err == nil {
		t.Fatal("expected error for failing command, got nil")
	}
}

func TestShellRunner_Integration_WithDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file in the temp directory
	testFile := filepath.Join(tmpDir, "hello.txt")
	if err := os.WriteFile(testFile, []byte("world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	runner := ShellRunner{}

	// Run cat in the temp directory - use relative path to verify dir works
	output, err := runner.Run("cat hello.txt", tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	trimmed := strings.TrimSpace(output)
	if trimmed != "world" {
		t.Errorf("expected 'world', got '%s'", trimmed)
	}
}

func TestShellRunner_Integration_WithExecute(t *testing.T) {
	runner := ShellRunner{}

	commands := dirCommands("", []config.Command{
		{Cmd: "echo step1"},
		{Cmd: "echo step2"},
	})

	output := captureLog(func() {
		Execute(runner, "integration-pkg", "int-001", commands)
	})

	if !strings.Contains(output, "step1") {
		t.Errorf("expected output to contain 'step1', got:\n%s", output)
	}
	if !strings.Contains(output, "step2") {
		t.Errorf("expected output to contain 'step2', got:\n%s", output)
	}
	if !strings.Contains(output, "command 1 of 2 succeeded") {
		t.Errorf("expected success log, got:\n%s", output)
	}
}
