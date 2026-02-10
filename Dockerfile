# ── Build stage ──────────────────────────────────────────────
FROM golang:1.24-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /steakpie ./cmd/steakpie

# ── Minimal runtime ────────────────────────────────────────
# Debian slim gives us bash (needed for `bash -lc` command execution).
# If you don't need the Docker socket, consider adding:
#   RUN useradd -r steakpie && chown steakpie /app
#   USER steakpie
FROM debian:bookworm-slim AS minimal

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /steakpie /usr/local/bin/steakpie

WORKDIR /app
EXPOSE 3142
ENTRYPOINT ["steakpie"]

# ── Docker runtime ──────────────────────────────────────────
# Extends minimal with docker CLI + compose plugin, for the
# common case of running `docker compose` commands from hooks.
FROM minimal AS docker

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      curl \
      gnupg \
 && curl -fsSL https://download.docker.com/linux/debian/gpg \
      | gpg --dearmor -o /usr/share/keyrings/docker.gpg \
 && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker.gpg] \
      https://download.docker.com/linux/debian bookworm stable" \
      > /etc/apt/sources.list.d/docker.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
      docker-ce-cli \
      docker-compose-plugin \
 && apt-get purge -y curl gnupg \
 && apt-get autoremove -y \
 && rm -rf /var/lib/apt/lists/*
