# ADR-0009: Backups are an opt-in integration

- **Status:** Accepted — 2026-06-15. Amended 2026-06-18: backups is now a
  catalog **integration** (ADR-0014), off by default, rather than a doc-only
  recipe -- but the isolation properties below are unchanged.

## Context
The agent works inside its bind-mounted home (`agent/`) and can delete or
overwrite files there, including uncommitted work. A safety net is worth having,
but hako's default should stay lean and "clone-and-up". A backup is also only
meaningful if the agent cannot reach (and wipe) it, and the box has `sudo`.

## Decision
Ship backups as an **opt-in integration** (`integrations/backups/`), off by
default and not in the base `compose.yaml`. It is **sidecar-only -- no skill** (so
the agent has no command to reach backups), with one `interval` setting. The
sidecar:
- mounts `./agent` read-only and writes to `./backups/`, which the main `hako`
  service does **not** mount (so the agent has no path to the repo or its key);
- uses **restic** (official image) for content-defined chunk dedup + compression,
  so a churning/reinstalled `node_modules` costs only changed chunks;
- generates its encryption key locally on first run into `./backups/` (never
  committed or shipped), keeping the "no secrets in the repo" property.

Enable in `hako.toml` (`[integrations.backups] enabled = true`) or via
`hako configure`; the assembler then composes the sidecar (`docs/backups.md`).

## Why
Default-on adds an always-running container and complexity most clones will not
want. The integration catalog (ADR-0014) gives exactly the property we want:
in-repo but dormant, toggled by config, with no extra surface when off -- better
than the original doc-only recipe because the wiring is real and exact (not prose
a human re-types), which matters since the isolation is security-critical: an
improvised setup might mount the backup dir into the main container and defeat
the point. Making it sidecar-only (no skill) keeps the repo unreachable by the
agent by construction.

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
