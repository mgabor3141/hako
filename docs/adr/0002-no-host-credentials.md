# ADR-0002: The agent holds no host credentials

- **Status:** Accepted — 2026-06-14

## Context
What is the security boundary for a sandboxed agent that can run arbitrary code?

## Decision
The container holds **zero host credentials**. The boundary is credential
*absence*, not behaviour restriction. Anything significant goes through the
(future) broker, which holds the upstream creds and is reachable only over a
private network with no host ports.

## Consequences
hako is safe to fork and publish. The agent can do whatever it likes locally but
can't reach real systems without the broker. Because only the agent can talk to
the broker, the broker needs no inbound auth.
