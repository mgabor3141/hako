# Production readiness

hako is validated end to end (pi + container + gmux + the MCP gateway + the
approval gate + the age vault + the integration catalog, against the GitHub
hosted MCP and a mock). It is a working prototype, not yet something to point at
real credentials and forget about. This is the honest list of placeholders,
known rough edges, and one-time steps to close before that.

## Before you trust it with real credentials

- **Revoke the tracer PAT and re-seal under a strong passphrase.** The dev vault
  was sealed with the throwaway passphrase `[redacted]` and a read-only token.
  `hako seal github` a real secret under a real passphrase; revoke the old PAT.
- **The secret reaches mcp-proxy as a process env var.** `hako unlock` decrypts
  on the host (locked buffer) and pipes the env in; the gateway sources it, so
  the value is in the mcp-proxy process environment. It is *not* in
  `docker inspect` (the compose env is blank) but it is readable via
  `/proc/<pid>/environ` by root inside the gateway. Closing this needs an
  mcp-proxy change to read the secret from a file/fd instead of the environment.
  (memguard locks the host-side blob; parsed Go string values still transit the
  heap -- a Go-language limitation.)

## The gateway

- **`callHook` lives in a fork.** The gateway image is pinned by `@sha256` digest
  (ADR-0008), built per commit from the `mcp-proxy` fork's `main`. Upstreaming
  `callHook` would remove the fork dependency, but isn't required.
- **No agent <-> gateway auth.** The boundary is the private compose network
  (`MCP_GATEWAY_TOKEN` defaults empty). Intentional, but know it: anything in the
  agent container can call the gateway.

## Integrations / composability

- **websearch ships a mock.** The bundled sidecar is `hashicorp/http-echo`
  returning two canned results -- it demonstrates the sidecar + typed-settings
  mechanism, it does not search. Point `url` at a real endpoint (and adjust the
  skill's response parsing) or wire a real backend (e.g. SearXNG) before relying
  on it.

## Approval UX

- The approval prompt appears as a **new gmux session you have to find by hand**:
  no push notification, no one-click dismiss. Unanswered requests **fail closed**
  after a TTL (900s). A persistent approvals view (and a gmux notify primitive)
  are the intended improvements (ADR-0010).

## Delivery

- The launcher ships per-commit **hash releases** (CI builds + publishes archives
  + checksums on every launcher change to `main`); the bootstrap pins a SHA and
  verifies the download. Bumping the pin is a committed `HAKO_VERSION` +
  `checksums.txt` diff. mac/WSL2 and arm64 builds are produced but unverified on
  those hosts.
- The `configure` TUI's **seal path** uses the standard `tea.ExecProcess` pattern
  but has not been driven through a full interactive session in CI.

## Tested against

- One real upstream (GitHub's hosted MCP) via a read-only token. Other MCP
  servers, other OSes (mac/WSL2 are designed-for but unverified), and arm64 have
  not been exercised.
