NOTE: NOT QUITE READY FOR PUBLIC USE


# steakpie


A really simple CI/CD server.

It listens for webhooks from Github, and when it receives one for a configured repository it runs the commands that you've specified. 

No abstractions, no elaborate interfaces.  A single binary, a single basic config file and you're off.

Initially designed to work with docker and docker-compose but in theory you can do whatever you want. The only current limitation is that it only listens for `Registry packages` webhook payloads.


## Quick start

1. Create a webhook on a repository you want to have deployed when a package is updated.
1. Download the binary to your server. `wget https://github.com/treejamie/steakpie/releases/download/v0.0.4/steakpie-0.0.4-arm64§
2. Maybe link it becasue your OCD won't let you run a binary with a messy name - `ln -s steakpie-0.0.4-arm64 steakpie`
3. Make it executable §chmod +x `steakpie-0.0.4-arm64`
4. Make sure you've got `$WEBHOOK_SECRET` in your environment and if you want to run on a port other than 3000, then set `$PORT`
5. Make a config file called config.yml or config.yaml
```bash
# name of your github repo
jamiec_ts
  - command to run when an webhook is sent
    - nested sub-command that runs only if parent succeeeds
  - paralell command that runs regardless of sibling succewss
```




## Why the package thing?

I was building a docker image as a build step when code was merged into main. I wanted to pull the latest image.

## Alternatives

### Coolify

Great, loads of options, but it was overkill. I just wanted to have my raspberry pi pull the latest version of an image when it was updated. I didn't want to lose a load of resource to running the central coolify platform on my pi, so I tried the cloud.  $5 is a steal, but I was scratching my head on how to get it to work and I kept hitting errors deploying things.

### Disco

Loved the idea, couldn't install it. Even with a fresh Ubuntu 24.04 I repeatedly failed. Aborted.

### K38's.

Life is too short for that much yaml.

### Manually updating.

No.




```
jamiec:  name of repo
  run:  # domain language.
    /foo/bar:  # this is the directory you want to cd'into
      - - echo "yay" # a command
        - echo "nested" # a nested command
      - cat foo.txt # a parallel command
    /another/dir:  # cd into another dir
      - ls -al # another command

```

You get to control what commands run and in what dirs. 

## TODO

- add setup and teardown keys
- support shells other than bash
- easy install / update script




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
