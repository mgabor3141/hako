# ADR-0001: Three repos, not a monorepo

- **Status:** Accepted — 2026-06-14

## Context
hako integrates pi + a container + gmux + (later) an MCP broker and agent-side
MCP skills. Do these live together?

## Decision
Three repos: **mcpeel** (token-efficient, hand-built CLIs over MCP, run by the
agent), **broker** (a fork of `TBXark/mcp-proxy` — the MCP gateway mcpeel
targets, holding the upstream creds), and **hako** (this repo: the opinionated
integration / distribution layer). hako **pins/references** the others; it does
not vendor them.

## Consequences
The pieces share a protocol (MCP) and a small convention, not code (TS vs Go),
and have different audiences/lifecycles — so they stay independently reusable
and upstream-trackable. hako is the thing people fork; the components are things
they track. Cost: cross-repo version pinning.
