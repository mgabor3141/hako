# ADR-0007: The broker runs host-local

- **Status:** Accepted — 2026-06-15 (resolves the host-or-sidecar ambiguity of
  the 2026-06-14 draft, which described both). `callHook` is built; not yet
  wired into hako.

## Context
Phase 2 goal: the agent reaches real tools without holding upstream credentials
(ADR-0002). The broker is a fork of `TBXark/mcp-proxy` — the **gateway** that
mcpeel's CLIs target (ADR-0001): it holds the upstream creds and exposes MCP
servers to the agent. The open question was *where it runs* relative to the
containerized agent: a credential-bearing sidecar, or host-local.

## Decision
In hako the broker runs **host-local** — a single static Go binary on the dev
machine, beside the credentials and credential helpers you already have
(`gh auth token`, cloud CLIs, keychains). It is **not** a hako container
service. Credentials are configured as a literal or a command; no
autodiscovery. The containerized agent reaches the broker over a **host-local
channel that hako never publishes to the LAN** — a loopback/bridge-bound port
or a bind-mounted unix socket (settled in the tracer bullet).

A per-server **`callHook`** gates chosen tool calls behind a command: the
request JSON arrives on stdin (+ `MCP_SERVER`/`MCP_TOOL`), exit 0 approves,
anything else denies (fail-closed) with a readable MCP error. The command is
**arbitrary and set in the broker's TOML**, so the approval channel is the
operator's choice — an OS notification, a TUI prompt, a webhook. Running
bare-metal is part of the point: a host process can raise a native desktop
notification, which a container cannot. hako ships example approval scripts,
not a baked-in channel.

## Consequences
Reuses host credential helpers directly (the main reason to run bare-metal) and
keeps the no-creds-in-the-agent property without a credential-bearing sidecar.
The same binary can still drop into a container for a homelab, but that is out
of scope for hako's default. Cost: "clone-and-up" no longer fully covers phase
2 — standing up the broker is a documented host-side setup step.

**Rejected alternative — credential-bearing sidecar:** clean container
isolation, but it cannot see your host credential helpers (you'd re-supply every
secret) and cannot raise OS notifications. Revisit only if we ever want a
zero-host-install distribution.
