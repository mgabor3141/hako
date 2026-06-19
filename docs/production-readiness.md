# Production readiness

hako is validated end to end (pi + container + gmux + the MCP gateway + the
approval gate + the age vault + the integration catalog, against the GitHub
hosted MCP and a mock). It is a working prototype, not yet something to point at
real credentials and forget about. This is the honest list of placeholders,
known rough edges, and one-time steps to close before that.

## Before you trust it with real credentials

- **Set up real credentials under a strong passphrase.** `hako auth github` with your
  own token and a real passphrase. (Development used a throwaway, read-only token
  under a throwaway passphrase, both since retired/revoked.)
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

- The launcher is **built from source** by the bootstrap (host Go, else a
  digest-pinned `golang` container), cached by source hash -- no release
  pipeline, no downloaded binary. The container path is verified on Linux/amd64;
  mac/WSL2 and arm64 are designed-for but unverified on those hosts.
- The `configure` TUI's **seal path** uses the standard `tea.ExecProcess` pattern
  but has not been driven through a full interactive session in CI.

## Tested against

- One real upstream (GitHub's hosted MCP) via a read-only token. Other MCP
  servers, other OSes (mac/WSL2 are designed-for but unverified), and arm64 have
  not been exercised.
