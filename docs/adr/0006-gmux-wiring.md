# ADR-0006: gmux container wiring (bind address, port, token)

- **Status:** Accepted — 2026-06-14

## Context
gmux gives browser access to sessions and runs inside the container. gmuxd binds
`127.0.0.1` by default and authenticates every TCP connection with a bearer
token.

## Decision
- `GMUXD_LISTEN=0.0.0.0` — a loopback bind inside the container is unreachable
  via a published port, so gmuxd must bind all interfaces.
- Publish `127.0.0.1:8791:8790` — host port 8791 (not gmux's default 8790) so a
  gmux on the host can coexist; the UI reaches only this machine's browser.
- `CMD ["gmuxd", "run"]` — gmux's documented foreground entry point, under tini.
- **No `GMUXD_TOKEN` shipped** — a fixed token in a public repo is a committed
  secret. The random token persists in the mounted state per-clone; retrieve it
  with `gmuxd auth`.

## Consequences
Browser UI works out of the box, scoped to localhost and token-gated. Remote /
phone access is a deliberate opt-in (gmux's Tailscale listener), not the default.
gmux is baked as a pinned, checksum-verified release binary (OS tier, ADR-0008).
As the container's main process it can only be updated by rebuild — an in-place
swap is ephemeral and would drop the very session serving it.
