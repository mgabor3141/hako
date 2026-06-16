# ADR-0002: The agent holds no host credentials

- **Status:** Accepted — 2026-06-14 (network claim restored 2026-06-16 once the
  gateway landed as a container sidecar again; see ADR-0007)

## Context
What is the security boundary for a sandboxed agent that can run arbitrary code?

## Decision
The container holds **zero host credentials**. The boundary is credential
*absence* plus per-call approval, not behaviour restriction. Anything
significant goes through the gateway, which holds the upstream creds and runs as
a sidecar on a **private compose network with no host ports** (ADR-0007),
reachable only by the agent over service DNS. Sensitive calls are additionally
gated by an approval hook (ADR-0010).

## Consequences
hako is safe to fork and publish. The agent can do whatever it likes locally but
can't reach real systems without the gateway. Because only the agent is on the
gateway's private network and it is never published, the gateway needs no inbound
auth in the default single-user setup; a multi-tenant or homelab deployment
would add one.
