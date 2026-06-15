# ADR-0009: Backups are an opt-in documented add-on

- **Status:** Accepted — 2026-06-15

## Context
The agent works inside its bind-mounted home (`agent/`) and can delete or
overwrite files there, including uncommitted work. A safety net is worth having,
but hako's default should stay lean and "clone-and-up". A backup is also only
meaningful if the agent cannot reach (and wipe) it, and the box has `sudo`.

## Decision
Ship backups as an **opt-in, documented add-on** ([`docs/backups.md`](../backups.md)),
not in the default `compose.yaml`. The recipe is a **sidecar container** that:
- mounts `./agent` read-only and writes to `./backups/`, which the main `hako`
  service does **not** mount (so the agent has no path to the repo or its key);
- uses **restic** (official image) for content-defined chunk dedup + compression,
  so a churning/reinstalled `node_modules` costs only changed chunks;
- generates its encryption key locally on first run into `./backups/` (never
  committed or shipped), keeping the "no secrets in the repo" property.

Enable with a second compose file: `docker compose -f compose.yaml -f
compose.backup.yaml up -d`.

## Why
Default-on adds an always-running container and complexity most clones will not
want. A profile would keep it in-repo but dormant; a doc-only recipe keeps the
default surface minimal and fits hako's "configured by an agent" model — you tell
your agent to follow the doc. The recipe is exact (not prose) precisely because
the isolation is security-critical: an improvised setup might mount the backup
dir into the main container and defeat the point.

restic over rsync `--link-dest`: hardlink snapshots dedup at whole-file
granularity, so `node_modules` churn defeats them. restic over borg: restic ships
single static binaries for amd64 and arm64; borg's stable standalone binaries are
x86-64 only.

## Consequences
Off by default; users opt in per the doc. Restore is a command
(`restic restore` / `dump` / `mount`), not a plain `cp`. The backup repo is
root-owned on the host (the sidecar runs as root), which also keeps casual
deletion behind `sudo`. Offsite (S3/SFTP) is a documented upgrade by pointing the
repo at a remote backend, with credentials given to the sidecar only.
