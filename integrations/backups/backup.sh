#!/bin/sh
# hako backup sidecar: periodic deduplicated snapshots of the agent home into an
# isolated restic repo that the main container cannot reach.
set -eu

: "${BACKUP_INTERVAL:=900}"   # seconds between snapshots

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
