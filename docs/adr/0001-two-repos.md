# ADR-0001: Two repos — hako (with the adapters) and a gateway fork

- **Status:** Accepted — 2026-06-14 (amended 2026-06-16: the agent-side MCP CLI
  adapters, formerly the standalone "mcpeel" repo, are now inlined into hako;
  was "three repos". See ADR-0013.)

## Context
hako integrates pi + a container + gmux + an MCP gateway + agent-side MCP CLI
adapters (the `github` CLI and friends). Do these live together?

## Decision
**Two repos:**
- **hako** (this repo) — the opinionated integration / distribution layer and
  the **reference implementation**. It now also **contains the agent-side MCP
  CLI adapters** in-tree (the token-efficient CLIs over MCP, formerly the
  standalone "mcpeel"), since they co-evolve and are co-tested here.
- **gateway** — a fork of `TBXark/mcp-proxy` (the MCP endpoint the adapters
  target, holding the upstream creds). hako **pins/references** it; does not
  vendor it.

The gateway stays separate because it is a genuine **upstream fork** (Go, tracks
and PRs back to TBXark) with its own lifecycle. The adapters were split off
speculatively before they had any external adopters; folding them back in
matches reality (tested here, shaped by hako sessions) without losing reuse —
they remain plain TS skills, standalone-runnable, and extractable later if they
earn it.

## Consequences
The adapters share hako's lifecycle, CI, and fork-and-`git pull` opinion model
(ADR-0005); no cross-repo pin for them. The gateway stays independently
trackable. hako is the thing people fork; the gateway is the thing they track.
Costs: hako's repo gains TypeScript (and the adapters' tests in CI), and the
adapters are less discoverable as a standalone library — acceptable, since they
have no external adopters yet and stay extractable.
