# Backups (optional integration)

hako does **not** back up by default. The agent works inside its home
(`agent/`), and a buggy or careless command can delete or overwrite files there.
If you want a safety net, enable the **`backups` integration**: a sidecar that
takes periodic, deduplicated, encrypted [restic](https://restic.net) snapshots of
the home into a repo the agent **cannot reach**.

```toml
# hako.toml  (or run `hako configure`)
[integrations.backups]
enabled = true
# interval = 900   # seconds between snapshots
```

Then `hako up` (or `docker compose ... up`) starts the sidecar.

## How it stays safe

The one rule that makes this worth anything: **the backup repo is never mounted
into the main `hako` container.** The sidecar mounts `./agent` read-only and
writes snapshots to `./backups/`, which the `hako` service does not mount. So
even though the agent has `sudo` in its container, it has no path to the backup
repo and cannot wipe its own history. The integration ships **no skill** either,
so the agent has no command to touch backups -- it is infrastructure, not a tool.

Why restic: content-defined chunking means a churning or reinstalled
`node_modules` only costs the chunks that actually changed (deduped, then
compressed), so backups stay small. It also encrypts at rest. The key is
**generated locally on first run** and lives only in `./backups/` -- never
committed or shipped, so hako keeps its "no secrets in the repo" property. The
image is pinned by digest (ADR-0008).

## Verify

```sh
docker compose -f compose.yaml -f integrations/backups/compose.yaml logs -f backup
docker compose -f compose.yaml -f integrations/backups/compose.yaml \
  run --rm --entrypoint restic backup snapshots
```

## Restore

```sh
base="-f compose.yaml -f integrations/backups/compose.yaml"

# list snapshots
docker compose $base run --rm --entrypoint restic backup snapshots

# restore the latest snapshot into ./restore on the host
mkdir -p restore
docker compose $base run --rm -v "$PWD/restore:/restore" \
  --entrypoint restic backup restore latest --target /restore

# recover a single file to stdout
docker compose $base run --rm --entrypoint restic backup \
  dump latest /src/path/to/file > recovered-file
```

## Tuning

- **Frequency / loss window:** the `interval` setting (seconds). Lower = smaller
  worst case if the agent deletes something between snapshots.
- **Retention / excludes:** the `restic forget --keep-*` flags and `--exclude`
  list in `integrations/backups/backup.sh` (edit directly; hako is fork-and-pull).
- **Offsite:** restic supports S3/SFTP/REST backends, but those need credentials
  delivered to the sidecar -- the vault only feeds the gateway today, so offsite
  is not wired yet (see `docs/production-readiness.md`).
