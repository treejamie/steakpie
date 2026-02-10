Prompt: .prompts/14.md

# Docker Support for Steakpie

## Context

Steakpie is a single Go binary CI/CD server. To make it easy for others to experiment with, we're adding Docker support. Rather than requiring users to build from source or download platform-specific binaries, a Dockerfile + Compose file lets them `docker compose up` and go.

## Key Constraints

- The binary executes commands via `bash -lc` (login shell), so **bash must exist** in the runtime image. Truly distroless won't work.
- Uses `modernc.org/sqlite` (pure Go) — no CGo needed, so `CGO_ENABLED=0` produces a static binary.
- Runtime needs: `WEBHOOK_SECRET` env var, `config.yml` in working dir, optional `PORT` (default 3142) and `DB_PATH` (default `db.sqlite`).

## Plan

### Step 1: Change default port to 3142

Update the Go code default from 3000 to 3142.

**File:** `cmd/steakpie/main.go` (line 72: `port = "3000"` → `port = "3142"`)

### Step 2: Create `.dockerignore`

Keeps the build context small — only Go source, `go.mod`, and `go.sum` need to reach the builder.

```
steakpie
*.sqlite*
.git
.github
.gitignore
.claude
.prompts
config.yaml
config.yml
testdata
*.md
version.txt
.DS_Store
```

**File:** `.dockerignore` (new)

### Step 3: Create `Dockerfile` with two named targets

Single multi-stage Dockerfile with three stages:

1. **`builder`** — `golang:1.24-bookworm`, builds the binary with `CGO_ENABLED=0`
2. **`minimal`** — `debian:bookworm-slim` + binary. Just bash and the binary (~45MB). For users running non-Docker commands or extending the image themselves.
3. **`docker`** — Extends `minimal`, adds `docker-ce-cli` + `docker-compose-plugin` (~100MB). For the common use case of running `docker compose` commands.

Usage:
- `docker build --target minimal .`
- `docker build --target docker .`

Both targets: `WORKDIR /app`, `EXPOSE 3142`, `ENTRYPOINT ["steakpie"]`.

No non-root user by default — if mounting the Docker socket, the container already has root-equivalent host access, making a non-root user security theatre. Commented note for users who don't need the socket.

**File:** `Dockerfile` (new)

### Step 4: Create `docker-compose.yml`

Default compose file using the `docker` target (the common case):

- Builds from local Dockerfile with `target: docker`
- `WEBHOOK_SECRET` from env/`.env` file
- `DB_PATH=/app/data/db.sqlite` for persistence in named volume
- Mounts: `config.yml:ro`, named volume for data, Docker socket
- `restart: unless-stopped`
- Port 3142

**File:** `docker-compose.yml` (new)

### Step 5: Commit and PR

Commit all files, create PR.

## Verification

1. `go test ./...` — existing tests still pass with new default port
2. `docker build --target minimal -t steakpie:minimal .` — builds successfully
3. `docker build --target docker -t steakpie:docker .` — builds successfully
4. `docker compose up` (with a valid `config.yml` and `WEBHOOK_SECRET` set) — container starts and listens on port 3142
