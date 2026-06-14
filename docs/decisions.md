# Architecture decisions

Condensed record of the choices behind hako and its sibling projects, so the
rationale survives outside chat history. Newest context at the bottom.

## 1. Three repos, not a monorepo

- **mcpeel** — agent-side MCP CLIs, distributed as a skills repo (its own ADRs
  mandate "the repo *is* the skills", no installer/Docker). Stays separate.
- **broker** — server-side MCP gateway (a fork of `TBXark/mcp-proxy`). Reusable
  with any MCP client.
- **hako** — this repo: the opinionated integration/distribution layer.

Why not a monorepo: the pieces share a *protocol* (MCP) + a tiny convention, not
*code* (TypeScript vs Go). Different languages, audiences, lifecycles, and
distribution models. hako **pins/references** the others; it does not vendor
them. Forkability is the deciding factor — hako is the thing people fork; the
components are things they track upstream.

## 2. The agent holds no credentials

The security boundary is **credential absence**, not behavior restriction. The
agent can do whatever it wants; it just has no tokens, so it can't do anything
significant without going through the (future) broker. The broker holds upstream
credentials; the agent reaches it over a private network with no host ports
(so only the agent can talk to it — which is also why the broker needs no
inbound auth).

## 3. Auth is pi's job; hako is auth-agnostic

hako ships no keys and assumes no provider. pi owns authentication
(bring-your-own, any provider), done once inside the container and persisted in
the home volume. This keeps hako safe to fork and publish.

## 4. The broker is a host-side artifact (planned)

Decided for the broker (separate repo, not built yet):
- A credential is configured as **a literal or a command** (e.g. `gh auth
  token`); no autodiscovery. The command runs where the creds live (the host).
- Therefore the broker should be a **single static binary** (Go), so it runs
  natively on a dev machine (no Python/Node, runs cred commands directly, no
  piping) *and* drops into a tiny container for homelab — same artifact.
- Built by forking `TBXark/mcp-proxy` (mark3labs/mcp-go), which already solves
  the hard part (aggregation/namespacing/session mapping).

## 5. `callHook` — gate tool calls behind a command (built, unpushed)

Added to the broker fork: a per-server `callHook` option (sibling to
`toolFilter`). Where `toolFilter` decides tool *visibility* at startup,
`callHook` runs *per invocation* as a mark3labs `ToolHandlerMiddleware`. The
configured command receives the request as JSON on stdin (+ `MCP_SERVER` /
`MCP_TOOL`); exit 0 approves, non-zero/timeout/error denies (fail-closed).
Denials return an MCP `isError` result so the model gets a readable reason. This
is the generic mechanism for human-in-the-loop approval; the opinionated
channels (editor/notify/telegram) ship as example scripts, not baked in.

## 6. hako delivers config via chezmoi-apply-this-repo

The container bootstraps by applying this repo as its config. Customization =
fork hako and point the bootstrap at the fork. Reuses the existing
container/devbox/chezmoi machinery instead of inventing a new config format.

## Status (Phase 1)

- hako: scaffolding the opinionated pi + container + gmux setup. No MCP yet.
- broker `callHook`: implemented + tested locally on a fork branch, not pushed,
  no PR.
- MCP layer (broker host-binary + mcpeel wiring): designed, not built.
