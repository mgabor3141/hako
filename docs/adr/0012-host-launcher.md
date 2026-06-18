# ADR-0012: Host-side `hako` launcher

- **Status:** Accepted — 2026-06-16. The launcher is a **Go engine** in
  `launcher/` (up/down/restart/assemble/seal/unlock/ps/logs/shell/pi/auth/open +
  passthrough): reads integration manifests, assembles the stack, wraps
  compose, and does the **in-process vault** (`filippo.io/age` + locked memory;
  single global passphrase). The interim shell launcher is **retired**; root
  `./hako` is now the **bootstrap** (model C) — it builds the Go binary from
  source for now (needs Go) and will fetch a pinned, checksummed release once CI
  publishes them. The `hako`->`help` rename is done. The `configure` TUI
  (Phase B) is **built** (bubbletea: toggle integrations, set typed settings,
  seal secrets, writes hako.toml). **Phase C** (release pipeline) is in place:
  `.github/workflows/launcher.yml` builds + tests every commit and, on `main`,
  publishes a **hash release** (archives + checksums under a release tagged with
  the commit SHA -- commit-hash identity, no semver, matching the gateway's
  digest model in ADR-0008). When Go is absent the bootstrap downloads the
  pinned SHA's archive and sha256-verifies it against the committed
  `launcher/checksums.txt`.

## Context
The default secret model needs an interactive unseal, MCP servers need guided
setup, and a first-timer shouldn't have to know `docker compose`. Something on
the host has to orchestrate the human-facing flow.

## Decision
Ship a host-side **`hako` Go binary** as the front door. `hako up`:
- runs `docker compose up` (agent + gateway sidecar),
- opens the gmux dashboard in **Chromium `--app` mode** if available, else the
  default browser (lifting gmux's `cli/gmux/cmd/gmux/browser.go`),
- **unseals the vault** (masked passphrase prompt; the host decrypts in-process
  and feeds the gateway's tmpfs — ADR-0011),
- and provides **`hako configure`** (a TUI) to enable integrations and seal their
  secrets into the `age` vault, never on a command line.

`docker compose up` stays available for power users; `hako up` is the documented
entry point. Go is chosen to reuse gmux's browser code and ship one static
cross-platform artifact, matching the gmux/gateway pinned-binary ethos.

**Delivery (model C):** commit a tiny, readable `./hako` **bootstrap script**.
With a Go toolchain it builds the launcher from source; without one it downloads
the **pinned, checksummed** `hako` binary -- a **per-commit hash release** (CI-
built, no semver tags; matching the gateway digest model) -- into a local cache,
verifies the sha256, and execs it. `git pull` bumps the pin. So the only
committed artifact is a script you can read in seconds; the binary follows
ADR-0008.

To avoid a name clash, the in-container cheatsheet command is renamed
`hako` → **`help`** (a zsh function, human shell only).

## Consequences
`git clone … && cd hako && ./hako up` with no Go toolchain, no manual download,
no committed binary (dodging repo bloat, multi-platform commits, and the
"is-the-blob-the-source?" audit problem — provenance comes from a CI-built
hash release, with SLSA attestations available later). Masked passphrase input
reveals length to a shoulder-surfer — accepted for first-timer clarity. The
launcher is now a fourth pinned artifact to maintain alongside mise/gmux/ffmpeg/
gateway.
