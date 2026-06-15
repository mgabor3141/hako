# ADR-0002: The agent holds no host credentials

- **Status:** Accepted — 2026-06-14 (network claim reconciled with the
  host-local broker, 2026-06-15; see ADR-0007)

## Context
What is the security boundary for a sandboxed agent that can run arbitrary code?

## Decision
The container holds **zero host credentials**. The boundary is credential
*absence* plus per-call approval, not behaviour restriction. Anything
significant goes through the broker, which holds the upstream creds and runs
host-local (ADR-0007). The broker is reachable only over a host-local channel
that hako never publishes to the LAN, and sensitive calls are additionally
gated by an approval hook.

## Consequences
hako is safe to fork and publish. The agent can do whatever it likes locally but
can't reach real systems without the broker. Because the broker's channel is
local-only and unpublished, it needs no inbound auth in the default single-user
setup; a multi-tenant or homelab deployment would add one.
