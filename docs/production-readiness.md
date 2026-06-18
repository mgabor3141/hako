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

- **Pinned to a branch, not a digest.** `gateway/compose.gateway.yaml` builds the
  fork from `...mcp-proxy.git#feat/call-hook`. Pin by image digest (ADR-0008) and
  ideally upstream the `callHook` change before release.
- **No agent <-> gateway auth.** The boundary is the private compose network
  (`MCP_GATEWAY_TOKEN` defaults empty). Intentional, but know it: anything in the
  agent container can call the gateway.

## Integrations / composability

- **The gateway overlay is GitHub-specific.** `gateway/compose.gateway.yaml`
  hardcodes `GITHUB_MCP_URL`/`GITHUB_MCP_TOKEN` on the gateway and
  `MCP_GATEWAY_URL: .../github/` on the agent. Today github is the only
  gateway-backed integration, so it works -- but a *second* one would need its
  upstream URL/secret env and its per-integration gateway route generated from
  the manifest, not hand-wired in the shared overlay. The "compose any subset"
  promise (ADR-0014) is met for skills + sidecars, only partly for multiple
  gateway backends.
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
- **`HAKO_APPROVE_ALL=1` bypasses approval entirely** -- it exists for the mock
  and headless/CI runs. Never set it in production.

## Delivery

- **No tagged release yet.** The no-Go bootstrap path can't fetch a binary until
  the first `v*` tag is cut and `launcher/HAKO_VERSION` + `launcher/checksums.txt`
  are committed (see `launcher/README.md`). Until then the bootstrap requires Go.
- The `configure` TUI's **seal path** uses the standard `tea.ExecProcess` pattern
  but has not been driven through a full interactive session in CI.

## Tested against

- One real upstream (GitHub's hosted MCP) plus the in-repo mock. Other MCP
  servers, other OSes (mac/WSL2 are designed-for but unverified), and arm64 have
  not been exercised.
