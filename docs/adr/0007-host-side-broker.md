# ADR-0007: The broker is a remote-proxy container sidecar

- **Status:** Accepted — 2026-06-16 (supersedes the 2026-06-15 "host-local"
  revision, which itself resolved the 2026-06-14 host-or-sidecar draft — see the
  history note below). `callHook` is built; not yet wired into hako.

## Context
Phase 2 goal: the agent reaches real tools without holding upstream credentials
(ADR-0002). The broker is a fork of `TBXark/mcp-proxy` — the **gateway** that
mcpeel's CLIs target (ADR-0001): it holds the upstream creds and exposes MCP
servers to the agent. The open question: *where does it run* relative to the
containerized agent, and *what does it run*?

The deciding fact: mcp-proxy aggregates two kinds of servers — **stdio**
(spawns a subprocess like `npx …`/`uvx …`, needing host runtimes) and **remote**
(`url`, just proxies an HTTP endpoint with a token). mcpeel's CLIs target
**remote** MCP servers (e.g. the GitHub remote MCP). Remote-proxying needs only
outbound HTTPS — no node/python — so it containerizes into a tiny Go-only image.

## Decision
hako's broker is **scoped to remote-proxying** and runs as a **container
sidecar** on hako's private compose network, Docker-managed. The agent reaches
it by **service DNS** (`http://broker:PORT`) — no host ports, no host-gateway,
no cross-platform binding wrinkle. The broker needs **no inbound auth** by
default (only the agent is on that network); mcpeel's gateway token is therefore
**optional** (a patch makes it so).

A per-server **`callHook`** gates chosen tool calls (those in `requireFor`)
behind a command — see ADR-0010.

**Host-process mode is documented, not default**, for power users who need
**stdio** servers with their host runtimes, or native OS notifications without a
D-Bus mount. The same Go binary serves both.

## Consequences
Docker — the one runtime hako already requires — gives lifecycle, restart, and
cross-platform "it just works" for free, with a tiny distroless-ish image pinned
per ADR-0008. No host supervisor (we'd ruled out assuming host gmux/systemd).
Restores ADR-0002's "private network, no host ports" as literally accurate.
Cost: stdio MCP servers are out of scope for the default broker (they'd need
host runtimes or a heavy node+python image); credentials must be injected rather
than read from host helpers (ADR-0011).

**History:** we briefly decided host-local (to reuse host credential helpers and
raise OS notifications), then flipped back here once the remote-proxy scope made
clear the container is tiny and Docker solves the supervisor problem the
host-process mode created.

**Rejected alternative — host-local process:** reuses host runtimes/credential
helpers and native notifications, but reintroduces the supervisor problem (no
gmux/systemd assumption) and a cross-platform host-reachability wrinkle. Kept as
the documented power-user mode.
