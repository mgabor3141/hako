#!/bin/sh
# Approval gate -- runs inside the gateway container (ADR-0010).
#
# OS-agnostic by design: it does NOT know how the human is asked. It drops the
# request into the shared /tmp/approvals volume and blocks until a verdict file
# appears (or callHook times out -> fail-closed deny). A hako-side watcher turns
# each request into an interactive gmux y/N session (browser, persistent).
#
# Input:  request params JSON on stdin (`{"name":...,"arguments":{...}}`);
#         MCP_SERVER / MCP_TOOL in env.
# Output: exit 0 approves; any non-zero denies (fail-closed).
set -eu

params="$(cat)"
server="${MCP_SERVER:-?}"; tool="${MCP_TOOL:-?}"
dir="${HAKO_APPROVALS_DIR:-/tmp/approvals}"
log() { echo "[approve] server=$server tool=$tool $*" >&2; }

# never-allow policy, regardless of the human
if [ "$tool" = "merge_pull_request" ]; then
  log "DENY (policy: merges are disabled)"
  echo "merging is disabled by hako policy" >&2
  exit 1
fi

# shared dir must be writable by the (unprivileged) hako-side watcher too
mkdir -p "$dir" 2>/dev/null || true
chmod 0777 "$dir" 2>/dev/null || true

id="${tool}-$(date +%s)-$$"
# {"server":..,"params":{"name":..,"arguments":..}} written atomically so the
# watcher never reads a partial file.
printf '{"server":"%s","params":%s}\n' "$server" "$params" > "$dir/$id.json.part"
mv "$dir/$id.json.part" "$dir/$id.json"

# block until the watcher's session writes a verdict; callHook kills us at its
# configured timeout, which surfaces as a deny.
while [ ! -f "$dir/$id.verdict" ]; do sleep 1; done
verdict="$(cat "$dir/$id.verdict")"

case "$verdict" in
  approve) log "APPROVE (operator)"; exit 0 ;;
  *)       log "DENY (operator)"; echo "denied by operator" >&2; exit 1 ;;
esac
