# ADR-0012: Host-side `hako` launcher

- **Status:** Accepted — 2026-06-16. The launcher is a **Go engine** in
  `launcher/` (up/down/restart/assemble/seal/unlock/ps/logs/shell/pi/auth/open +
  passthrough): reads integration manifests, assembles the stack, wraps
  compose, and does the **in-process vault** (`filippo.io/age` + locked memory;
  single global passphrase). The interim shell launcher is **retired**; root
  `./hako` is now the **bootstrap** — it builds the launcher (host Go, else a
  digest-pinned `golang` container) and caches it by source hash. The
  `hako`->`help` rename is done. The `configure` TUI
  (Phase B) is **built** (bubbletea: toggle integrations, set typed settings,
  set up auth, writes hako.toml). **Delivery** is build-from-source, cached by
  source hash -- no release pipeline and no downloaded binary, so the binary
  never skews from the source you pulled.

## Context
The default secret model needs an interactive unlock, MCP servers need guided
setup, and a first-timer shouldn't have to know `docker compose`. Something on
the host has to orchestrate the human-facing flow.

## Decision
Ship a host-side **`hako` Go binary** as the front door. `hako up`:
- runs `docker compose up` (agent + gateway sidecar),
- opens the gmux dashboard in **Chromium `--app` mode** if available, else the
  default browser (lifting gmux's `cli/gmux/cmd/gmux/browser.go`),
- **unlocks the vault** (masked passphrase prompt; the host decrypts in-process
  and feeds the gateway's tmpfs — ADR-0011),
- and provides **`hako configure`** (a TUI) to enable integrations and set up their
  secrets into the `age` vault, never on a command line.

`docker compose up` stays available for power users; `hako up` is the documented
entry point. Go is chosen to reuse gmux's browser code and ship one static
cross-platform artifact, matching the gmux/gateway pinned-binary ethos.

**Delivery:** commit a tiny, readable `./hako` **bootstrap script** that builds
the launcher and caches the binary by a hash of the launcher source (rebuild
only on change), then execs it. With a host Go toolchain it builds directly;
without one it builds in a **digest-pinned `golang` container** -- Docker is
hako's one hard dependency, so this always works -- run as the host UID so the
binary isn't root-owned. No release pipeline and no downloaded binary: you always
run a build of the exact source in your checkout, so there's nothing to pin or
verify beyond the repo you cloned (the toolchain image is digest-pinned per
ADR-0008; deps are `go.sum`-pinned). The earlier model -- a CI-published,
checksum-pinned hash release for no-Go hosts -- was dropped: leaning on Docker
removes the release pipeline, the in-repo pin, and any binary-vs-source skew.

To avoid a name clash, the in-container cheatsheet command is renamed
`hako` → **`help`** (a zsh function, human shell only).

## Consequences
`git clone … && cd hako && ./hako up` with no Go toolchain, no manual download,
no committed binary (dodging repo bloat, multi-platform commits, and the
"is-the-blob-the-source?" audit problem — there's no blob; you build your own
checkout's source in a pinned toolchain). Masked passphrase input
reveals length to a shoulder-surfer — accepted for first-timer clarity. The
launcher itself is not a pinned artifact (it's built from in-repo source); only
its build toolchain image is pinned, alongside mise/gmux/ffmpeg/gateway.
