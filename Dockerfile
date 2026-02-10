# ── Build stage ──────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /steakpie ./cmd/steakpie

# ── Minimal runtime ────────────────────────────────────────
# Alpine + bash (needed for `bash -lc` command execution).
# If you don't need the Docker socket, consider adding:
#   RUN adduser -D steakpie && chown steakpie /app
#   USER steakpie
FROM alpine:3.21 AS minimal

RUN apk add --no-cache bash

COPY --from=builder /steakpie /usr/local/bin/steakpie

WORKDIR /app
EXPOSE 3142
ENTRYPOINT ["steakpie"]

# ── Docker runtime ──────────────────────────────────────────
# Extends minimal with docker CLI + compose plugin, for the
# common case of running `docker compose` commands from hooks.
FROM minimal AS docker

RUN apk add --no-cache docker-cli docker-cli-compose
