# Steak Pie

Simple CI/CD server — listens for GitHub webhooks, runs configured commands. Single Go binary, single config file.

## Stack
- Go
- Docker Compose (deployment on Raspberry Pi via Cloudflare Tunnel)

## Commands
- `go test ./...` — run all tests
- `go build -o steakpie` — build binary
- `GOOS=linux GOARCH=arm64 go build -o steakpie` — build for Pi

## Workflow
See @AGENTS.md for branching, commit, PR, and test rules.

## Important
- NEVER unstage files in .prompts/ or version.txt
- Tests MUST pass before moving to next plan step
- Fix failing tests by fixing implementation, not tests (ask permission to edit tests)
- PR titles MUST be conventional commit format (used for squash merge)