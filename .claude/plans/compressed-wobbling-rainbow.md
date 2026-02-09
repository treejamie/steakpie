# Plan: New config format with `run` key and directory-based execution

## Context

Commands configured in the YAML config were executing without `cd`-ing into the correct directory first, causing failures. The config format is being restructured to explicitly specify which directory each group of commands should run in. A new `run` key is introduced as domain language (with future keys like `set_up` and `teardown` planned but not for today).

**Old format:**
```yaml
jamiec:
  - docker compose down
  - docker compose up -d
```

**New format:**
```yaml
jamiec:
  run:
    /foo/bar:
      - - echo "yay"
        - echo "nested"
      - cat foo.txt
    /another/dir:
      - ls -al
```

Key structural changes:
- Package value changes from a flat command list to an object with a `run` key
- `run` maps directory paths to command lists
- Command nesting changes from mapping-based (`cmd: [children]`) to sequence-based (`- - parent\n  - child`)

## Files to modify

### 1. `internal/config/config.go` - Config types and parsing

**Type changes:**
```go
// PackageConfig holds the configuration for a single package.
type PackageConfig struct {
    Run map[string][]Command `yaml:"run"`
}

// Config maps package names to their configuration.
type Config map[string]PackageConfig
```

**Command.UnmarshalYAML changes:**
Replace `MappingNode` handling with `SequenceNode` handling. The nesting format changes from mapping-based (`cmd: [children]`) to sequence-based (`[parent, child1, child2]`):

- `ScalarNode` → simple command (unchanged)
- `SequenceNode` → first element must be a scalar (parent command string), remaining elements are children (recursive, so children can themselves be sequences for deeper nesting)
- Remove `MappingNode` case entirely — return error for unexpected node kinds

**GetCommands → GetRun:**
```go
func (c Config) GetRun(packageName string) map[string][]Command {
    pkg, ok := c[packageName]
    if !ok {
        return nil
    }
    return pkg.Run
}
```

### 2. `internal/executor/executor.go` - Runner interface and execution

**Runner interface change** - add `dir` parameter:
```go
type Runner interface {
    Run(cmd string, dir string) (output string, err error)
}
```

**ShellRunner** - use `exec.Command.Dir`:
```go
func (s ShellRunner) Run(cmd string, dir string) (string, error) {
    c := exec.Command("sh", "-c", cmd)
    if dir != "" {
        c.Dir = dir
    }
    out, err := c.CombinedOutput()
    return string(out), err
}
```

**Execute** - accept `map[string][]Command`, iterate directories:
```go
func Execute(runner Runner, packageName, deliveryID string, dirCommands map[string][]Command) {
    log.Printf("start webhook for %s received with id: %s", packageName, deliveryID)
    for dir, commands := range dirCommands {
        log.Printf("executing in directory: %s", dir)
        executeLevel(runner, dir, commands)
    }
    log.Printf("end webhook for %s with id: %s", packageName, deliveryID)
}
```

**executeLevel** - pass `dir` through to `runner.Run`:
```go
func executeLevel(runner Runner, dir string, commands []Command) {
    // same logic, but runner.Run(cmd.Cmd, dir) instead of runner.Run(cmd.Cmd)
}
```

### 3. `internal/webhook/handler.go` - Use new config API

Change the command lookup section:
```go
packageName := event.RegistryPackage.Name
dirCommands := cfg.GetRun(packageName)

if len(dirCommands) > 0 {
    go executor.Execute(runner, packageName, deliveryID, dirCommands)
} else {
    log.Printf("No commands configured for package %s", packageName)
}
```

### 4. `internal/config/config_test.go` - Update all test YAML fixtures

Update all test YAML to use the new format. Every test that constructs YAML or `Config` literals needs updating.

### 5. `internal/executor/executor_test.go` - Update MockRunner and tests

- `MockRunner.Run` gains `dir string` parameter
- All test command construction changes to use `map[string][]Command`
- Add test: commands in different directories use the correct dir
- Add test: ShellRunner integration test that verifies directory-based execution (e.g., create a temp dir with a file, run `cat` in that dir)

### 6. `internal/webhook/handler_test.go` - Update testConfig

Change `testConfig` from `config.Config` with flat command lists to the new `PackageConfig` structure with `run` and directory mappings.

### 7. `cmd/steakpie/main_test.go` - Update test YAML fixtures

Update any inline YAML used in tests to match the new format.

### 8. `config.yaml` and `test-config.yaml` - Update example configs

Update to new format.

## Execution order

1. Update `internal/config/config.go` (types + parsing)
2. Update `internal/executor/executor.go` (Runner interface + Execute)
3. Update `internal/webhook/handler.go` (use new API)
4. Update `internal/config/config_test.go`
5. Update `internal/executor/executor_test.go`
6. Update `internal/webhook/handler_test.go`
7. Update `cmd/steakpie/main_test.go`
8. Update `config.yaml` and `test-config.yaml`
9. Run all tests, fix any issues

## Verification

- `go test ./...` — all tests pass
- `go vet ./...` — no issues
- Manual review of config.yaml to confirm it matches the new schema
