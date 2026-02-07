# steakpie

A really simple CI/CD server.

It listens for webhooks from Github, and when it receives one for a configured repository it runs the commands that you've specified. 

No abstractions, no elaborate interfaces.  A single binary, a single basic config file and you're off.


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
