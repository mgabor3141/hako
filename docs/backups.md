# Backups (optional add-on)

hako does **not** back up by default. The agent works inside its home
(`agent/`), and a buggy or careless command can delete or overwrite files
there. If you want a safety net, this guide sets up a small **backup sidecar**:
a separate container that takes periodic, deduplicated, encrypted snapshots of
the home into a repo the agent **cannot reach**.

You can hand this whole file to your agent: *"set up backups as described in
`docs/backups.md`."*

## How it stays safe

The one rule that makes this worth anything: **the backup target is never
mounted into the main `hako` container.** The sidecar mounts `./agent` read-only
and writes snapshots to `./backups/`, which the `hako` service does not mount. So
even though the agent has `sudo` in its container, it has no path to the backup
repo and cannot wipe its own history.

Why [restic](https://restic.net): content-defined chunking means a churning or
reinstalled `node_modules` only costs the chunks that actually changed (deduped
repo-wide, then compressed), so backups stay small even as projects grow. It also
encrypts at rest. The encryption key is **generated locally on first run** and
lives only in `./backups/` — it is never committed or shipped, so hako keeps its
"no secrets in the repo" property. We use restic's official image
(`restic/restic`, a ~68 MB busybox image with restic and a shell), so there is
nothing to build.

## Setup

Create two files.

**1. `container/backup.sh`** — the snapshot loop:

```sh
#!/bin/sh
# hako backup sidecar: periodic deduplicated snapshots of the agent home into an
# isolated restic repo that the main container cannot reach.
set -eu

: "${BACKUP_INTERVAL:=900}"   # seconds between snapshots (default 15 min)

# Generate a random repo key on first run. It lives only here (gitignored,
# unreachable by the main container), so hako still ships no secrets.
if [ ! -s "$RESTIC_PASSWORD_FILE" ]; then
  head -c 32 /dev/urandom | base64 > "$RESTIC_PASSWORD_FILE"
  chmod 600 "$RESTIC_PASSWORD_FILE"
fi

# Initialize the repo once.
restic cat config >/dev/null 2>&1 || restic init

echo "hako-backup: snapshotting /src every ${BACKUP_INTERVAL}s -> ${RESTIC_REPOSITORY}"
while true; do
  # Reinstallable / derived dirs are excluded: they bloat scans and restore for
  # no benefit. Add your own (target, dist, build, .next, ...) as needed.
  restic backup /src --tag hako \
    --exclude='.local' --exclude='.cache' \
    --exclude='.bun'   --exclude='.npm'   \
    --exclude='node_modules' --exclude='.venv' --exclude='venv' \
    || echo "hako-backup: backup failed, will retry next cycle"

  restic forget --tag hako \
    --keep-last 48 --keep-hourly 24 --keep-daily 14 --keep-weekly 8 \
    --prune || true

  sleep "$BACKUP_INTERVAL"
done
```

**2. `compose.backup.yaml`** — the sidecar service (kept out of the default
`compose.yaml` so it stays opt-in):

```yaml
# hako backup add-on. Enable with:
#   docker compose -f compose.yaml -f compose.backup.yaml up -d
# ./backups is deliberately NOT mounted into the `hako` service, so the agent
# cannot reach or delete its own backups.
services:
  backup:
    image: restic/restic:0.18.1      # pin by @sha256:... for stricter integrity
    container_name: hako-backup
    restart: unless-stopped
    entrypoint: ["/bin/sh", "/backup.sh"]
    environment:
      BACKUP_INTERVAL: "900"
      RESTIC_REPOSITORY: /backups/repo
      RESTIC_PASSWORD_FILE: /backups/.restic-pass
    volumes:
      - ./agent:/src:ro                    # read-only source
      - ./backups:/backups                 # isolated repo (hako does not mount this)
      - ./container/backup.sh:/backup.sh:ro
```

Then ignore the backup dir and turn it on:

```sh
echo '/backups/' >> .gitignore
docker compose -f compose.yaml -f compose.backup.yaml up -d
```

To keep the flags out of every command, you can instead
`export COMPOSE_FILE=compose.yaml:compose.backup.yaml` in your shell.

## Verify

```sh
docker compose -f compose.yaml -f compose.backup.yaml logs -f backup   # watch it run
docker compose -f compose.yaml -f compose.backup.yaml run --rm --entrypoint restic backup snapshots
```

## Restore

The repo and key are read from the service's environment, so these are short:

```sh
# List snapshots
docker compose -f compose.yaml -f compose.backup.yaml run --rm --entrypoint restic backup snapshots

# Restore the latest snapshot into ./restore on the host
mkdir -p restore
docker compose -f compose.yaml -f compose.backup.yaml run --rm \
  -v "$PWD/restore:/restore" --entrypoint restic backup restore latest --target /restore

# Recover a single file to stdout
docker compose -f compose.yaml -f compose.backup.yaml run --rm --entrypoint restic backup \
  dump latest /src/path/to/file > recovered-file
```

## Tuning

- **Frequency / loss window:** `BACKUP_INTERVAL` (seconds). Lower = smaller worst
  case if the agent deletes something between snapshots.
- **Retention:** the `restic forget --keep-*` flags. Dedup makes generous
  retention cheap.
- **Excludes:** add `target`, `dist`, `build`, `.next`, large data dirs, etc.
- **Offsite:** point `RESTIC_REPOSITORY` at an S3/SFTP/REST backend (and supply
  its credentials to the **sidecar only**) to keep copies off the machine.
