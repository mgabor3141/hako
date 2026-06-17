# ADR-0012: Host-side `hako` launcher

- **Status:** Accepted — 2026-06-16. The launcher is a **Go engine** in
  `launcher/` (up/down/restart/assemble/seal/unlock/ps/logs/shell/pi/auth/open +
  passthrough, `--mock`): reads integration manifests, assembles the stack, wraps
  compose, and does the **in-process vault** (`filippo.io/age` + locked memory;
  single global passphrase). The interim shell launcher is **retired**; root
  `./hako` is now the **bootstrap** (model C) — it builds the Go binary from
  source for now (needs Go) and will fetch a pinned, checksummed release once CI
  publishes them. The `hako`->`help` rename is done. The `configure` TUI
  (Phase B) is **built** (bubbletea: toggle integrations, set typed settings,
  seal secrets, writes hako.toml). **Phase C** (release pipeline) is in place:
  `.goreleaser.yaml` + a tag-triggered CI workflow build pinned, checksummed
  binaries, and the bootstrap downloads + sha256-verifies them against the
  committed `launcher/checksums.txt` when Go is absent. The first tagged release
  (and committing the `HAKO_VERSION` + `checksums.txt` pin) is the remaining
  one-time step.

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
