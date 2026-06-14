# ADR-0007: Host-side MCP broker

- **Status:** Proposed — 2026-06-14 (not built/wired; lives in the broker repo, see ADR-0001)

## Context
Phase 2 goal: the agent reaches tools without holding upstream credentials
(ADR-0002).

## Decision
The broker is a **single static Go binary** (fork of `TBXark/mcp-proxy`), so the
same artifact runs natively on a dev host — executing credential commands like
`gh auth token` directly — and drops into a tiny container for a homelab.
Credentials are configured as a **literal or a command**, no autodiscovery.

A per-server **`callHook`** (built, unpushed) gates each tool call behind a
command: request JSON on stdin (+ `MCP_SERVER`/`MCP_TOOL`), exit 0 approves,
anything else denies (fail-closed) with a readable MCP error — the generic
human-in-the-loop approval mechanism.

## Consequences
Same binary for dev and homelab; no Python/Node runtime. Not yet integrated with
hako. Opinionated approval channels ship as example scripts, not baked in.
