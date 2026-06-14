# ADR-0006: gmux container wiring (bind address, port, token)

- **Status:** Accepted — 2026-06-14

## Context
gmux gives browser access to sessions and runs inside the container. gmuxd binds
`127.0.0.1` by default and authenticates every TCP connection with a bearer
token.

## Decision
- `GMUXD_LISTEN=0.0.0.0` — a loopback bind inside the container is unreachable
  via a published port, so gmuxd must bind all interfaces.
- Publish `127.0.0.1:8790:8790` — the UI reaches only this machine's browser.
- `CMD ["gmuxd", "run"]` — gmux's documented foreground entry point, under tini.
- **No `GMUXD_TOKEN` shipped** — a fixed token in a public repo is a committed
  secret. The random token persists in the mounted state per-clone; retrieve it
  with `gmuxd auth`.

## Consequences
Browser UI works out of the box, scoped to localhost and token-gated. Remote /
phone access is a deliberate opt-in (gmux's Tailscale listener), not the default.
