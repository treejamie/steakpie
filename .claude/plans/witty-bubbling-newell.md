# Fix: Nested command config on Pi

## Context
The `jamiec_ts` config on the Pi is missing the colon after the parent command, causing YAML to fold both lines into a single scalar string instead of parsing them as a nested parent/child structure.

**Pi config (broken):**
```yaml
jamiec_ts:
  - docker compose pull
    - doppler run -- docker compose up -d
```
YAML sees one scalar: `"docker compose pull - doppler run -- docker compose up -d"`

**Repo config (correct):**
```yaml
jamiec_ts:
  - docker compose pull:
      - doppler run -- docker compose up -d
```
YAML sees a mapping: parent `docker compose pull` with child `doppler run -- docker compose up -d`

## Resolution
Not a code bug — config formatting issue on the Pi. Fix by adding the colon (`:`) after `docker compose pull` and indenting the child list (6 spaces or 2 extra beyond parent).

No code changes needed.
