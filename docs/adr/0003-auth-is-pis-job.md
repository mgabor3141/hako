# ADR-0003: Auth is pi's job; hako is auth-agnostic

- **Status:** Accepted — 2026-06-14

## Context
Which model provider / API keys does hako assume?

## Decision
None. pi owns authentication (bring-your-own, any provider), done once inside
the container and persisted in the mounted home.

## Consequences
hako ships no keys and is safe to publish. The user authenticates pi after the
first `docker compose up`; the credential lives in the user's clone, gitignored,
never in the image.
