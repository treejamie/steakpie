# steakpie

A GitHub webhook handler for registry package events.

## Running

### Prerequisites

Set the required environment variable:
```bash
export WEBHOOK_SECRET=your-github-webhook-secret
```

### Start the server

```bash
./steakpie
```

The server will start on port 3000 by default.

### Custom port

```bash
PORT=3000 ./steakpie
```

### Example

```bash
WEBHOOK_SECRET=my-secret-key ./steakpie
```

## Building

```bash
go build -o steakpie ./cmd/steakpie
```
