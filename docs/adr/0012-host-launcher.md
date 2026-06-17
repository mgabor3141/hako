# ADR-0012: Host-side `hako` launcher

- **Status:** Accepted — 2026-06-16. An interim POSIX shell `./hako`
  (`up`/`down`/`restart`/`seal`/`unlock`/`ps`/`logs`/`shell`/`pi`/`auth`/`open` +
  passthrough, `--mock`) is built; `up` auto-detects the vault and unseals it
  (masked passphrase piped to the gateway), `seal` encrypts a secret into the
  vault (the secret-entry half of the wizard below), `unlock` re-does the unseal
  after a gateway restart, and the `hako`->`help` rename is done. The Go engine
  now exists in `launcher/` at parity (Phase A: reads integration manifests,
  assembles the stack, wraps compose + vault); the `configure` TUI, the
  in-process vault (age library + locked memory), and bootstrap delivery remain.

## Context
The default secret model needs an interactive unseal, MCP servers need guided
setup, and a first-timer shouldn't have to know `docker compose`. Something on
the host has to orchestrate the human-facing flow.

## Decision
Ship a host-side **`hako` Go binary** as the front door. `hako up`:
- runs `docker compose up` (agent + gateway sidecar),
- opens the gmux dashboard in **Chromium `--app` mode** if available, else the
  default browser (lifting gmux's `cli/gmux/cmd/gmux/browser.go`),
- **unseals the vault** (masked passphrase prompt, piped to the gateway over
  stdin — ADR-0011),
- and provides an **interactive wizard** to set up MCP servers and their auth
  (storing secrets in the `age` vault, never on a command line).

`docker compose up` stays available for power users; `hako up` is the documented
entry point. Go is chosen to reuse gmux's browser code and ship one static
cross-platform artifact, matching the gmux/gateway pinned-binary ethos.

**Delivery (model C):** commit a tiny, readable `./hako` **bootstrap script**;
on first run it detects OS/arch, downloads the **pinned, checksummed** `hako`
release binary (goreleaser, CI-built from a tag) into a local cache, verifies
the sha256, and execs it. `git pull` bumps the pin. So the only committed
artifact is a script you can read in seconds; the binary follows ADR-0008.

To avoid a name clash, the in-container cheatsheet command is renamed
`hako` → **`help`** (a zsh function, human shell only).

## Consequences
`git clone … && cd hako && ./hako up` with no Go toolchain, no manual download,
no committed binary (dodging repo bloat, multi-platform commits, and the
"is-the-blob-the-source?" audit problem — provenance comes from a CI-built
tagged release, with SLSA attestations available later). Masked passphrase input
reveals length to a shoulder-surfer — accepted for first-timer clarity. The
launcher is now a fourth pinned artifact to maintain alongside mise/gmux/ffmpeg/
gateway.
