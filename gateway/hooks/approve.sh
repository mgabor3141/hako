#!/bin/sh
# callHook approval gate (runs inside the gateway container).
#
# Input:  the tool-call request JSON on stdin; MCP_SERVER / MCP_TOOL in env.
# Output: exit 0 approves the call; any non-zero exit denies it (fail-closed).
#         Whatever we print to stderr/stdout becomes the model-visible refusal.
#
# This is the simple, swappable default (ADR-0010). It shows the three policy
# branches — allow / deny / ask — driven by a per-tool table. The real
# interactive prompt (notify-send click-to-approve) swaps in for the "ask"
# branch later; for now "ask" honors HAKO_APPROVE_ALL so the loop is testable.
set -eu

req="$(cat)"
log() { echo "[approve] server=${MCP_SERVER:-?} tool=${MCP_TOOL:-?} $*" >&2; }

case "${MCP_TOOL:-}" in
  merge_pull_request)
    # deny: never allowed by policy
    log "DENY (policy: merges are disabled)"
    echo "merging is disabled by hako policy" >&2
    exit 1
    ;;
  *)
    # ask: needs operator approval. notify-send goes here; until then a toggle.
    if [ "${HAKO_APPROVE_ALL:-0}" = "1" ]; then
      log "APPROVE"
      exit 0
    fi
    log "DENY (awaiting approval; set HAKO_APPROVE_ALL=1 to approve)"
    echo "awaiting operator approval — none granted" >&2
    exit 1
    ;;
esac
