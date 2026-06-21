# Production readiness

hako is validated end to end (pi + container + gmux + the MCP gateway + the
approval gate + the age vault + the integration catalog) against GitHub's hosted
MCP, the websearch and webview sidecars, and an office document round-trip. It is
a working prototype, not yet something to point at real credentials and forget
about. This is the honest list of caveats, known rough edges, and one-time steps
to close before that.

## Before you trust it with real credentials

- **Set up real credentials under a strong passphrase.** `hako auth github` with
  your own token and a real passphrase. (Development used a throwaway, read-only
  token under a throwaway passphrase, both since retired/revoked.)
- **The secret reaches the gateway as a process env var.** `hako up` (and `hako
  unlock`) decrypts on the host (locked buffer) and pipes the env in; the gateway
  sources it, so the value lives in the mcp-proxy process environment. It is *not*
  in `docker inspect` (the compose env is blank) but it is readable via
  `/proc/<pid>/environ` by root inside the gateway container. Closing this needs
  an mcp-proxy change to read the secret from a file/fd instead of the
  environment. (memguard locks the host-side blob; parsed Go string values still
  transit the heap -- a Go-language limitation.)
- **The agent holds exactly one credential, on purpose.** Everything else is
  credential-absence-plus-approval. Git *push* is the one bounded exception
  (ADR-0015): a transport-only, human-provisioned, Contents-scoped PAT that lives
  in the bind-mounted home (`~/.git-credentials`). It is deliberately narrower
  than the gateway, opening/merging PRs is still approval-gated through the
  `github` tool, and `main` should be branch-protected so a push cannot land
  unreviewed.

## The gateway

- **`callHook` lives in a fork.** The gateway image is pinned by `@sha256` digest
  (ADR-0008), built per commit from the `mcp-proxy` fork's `main`. Upstreaming
  `callHook` would remove the fork dependency, but isn't required.
- **No agent <-> gateway auth.** The boundary is the private compose network
  (`MCP_GATEWAY_TOKEN` defaults empty). Intentional, but know it: anything in the
  agent container can call the gateway.

## Integrations / composability

- **websearch and webview are real sidecars you start.** websearch is a SearXNG
  instance; webview is crawl4ai. Both images are digest-pinned (ADR-0008).
  SearXNG's rate limiter is intentionally off for single-user local use -- fine
  on a private host, reconsider before exposing it.
- **office has no render/verify loop, by design.** It is pure-Python (python-docx
  / openpyxl / python-pptx): the agent builds documents structurally and verifies
  by re-reading them, but cannot *see* a rendered page (so e.g. a formula reads
  back as its literal string, not its computed value). Fidelity rendering would
  mean LibreOffice (~1 GB) or an unvetted binary, deliberately out of scope. This
  is a scope choice, not a placeholder, and its SKILL.md says so.

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
- **Pins are paired with a cadence -- once you switch it on.** Images and tools
  are digest/version-pinned (ADR-0008) and `renovate.json` is committed, so
  Renovate raises update PRs. Pinning only pays off with the cadence, so enable
  the Mend app (or a self-hosted `renovatebot/github-action`) on the repo.
- The `configure` TUI's **auth path** uses the standard `tea.ExecProcess` pattern
  but has not been driven through a full interactive session in CI.

## Tested against

- GitHub's hosted MCP via a read-only token; the websearch (SearXNG) and webview
  (crawl4ai) sidecars; and an office .docx/.xlsx/.pptx round-trip. Other MCP
  servers, other OSes (mac/WSL2 are designed-for but unverified), and arm64 have
  not been exercised.
