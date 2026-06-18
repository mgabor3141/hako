#!/bin/sh
# hako-side approval watcher -- runs in the hako container, where gmux lives
# (the gateway can't reach gmuxd's local socket across containers). Turns each
# request the gateway hook drops in /tmp/approvals into an interactive gmux y/N
# session (prompt.sh), and reaps stale artifacts by age. ADR-0010.
#
# Started by the entrypoint when HAKO_APPROVAL_WATCH points here (set by the
# gateway compose overlay).
set -eu
here="$(cd "$(dirname "$0")" && pwd)"
prompt="$here/prompt.sh"
dir="${HAKO_APPROVALS_DIR:-/tmp/approvals}"
ttl="${HAKO_APPROVALS_TTL:-900}"   # seconds before unanswered artifacts are reaped
mkdir -p "$dir" 2>/dev/null || true

while :; do
  # new requests -> a detached gmux session, cwd = the approvals dir
  for req in "$dir"/*.json; do
    [ -e "$req" ] || continue
    id="$(basename "$req" .json)"
    [ -f "$dir/$id.seen" ] && continue
    : > "$dir/$id.seen"
    ( cd "$dir" && gmux --no-attach "$prompt" "$id" ) > "$dir/$id.sid" 2>/dev/null || true
  done

  # age-based reap (covers timed-out, never-answered requests + old logs)
  now="$(date +%s)"
  for f in "$dir"/*.seen; do
    [ -e "$f" ] || continue
    mtime="$(stat -c %Y "$f" 2>/dev/null || echo "$now")"
    [ $((now - mtime)) -gt "$ttl" ] || continue
    id="$(basename "$f" .seen)"
    sid="$(cat "$dir/$id.sid" 2>/dev/null || true)"
    if [ -n "$sid" ]; then gmux --kill "$sid" >/dev/null 2>&1 || true; fi
    rm -f "$dir/$id.json" "$dir/$id.seen" "$dir/$id.sid" "$dir/$id.verdict"
  done

  sleep 1
done
