# ADR-0004: Two tiers — thin Debian OS image, mise-managed home toolchain

- **Status:** Accepted — 2026-06-15 (supersedes the devbox base, and the later "bake every tool into the image via apt + curl")

## Context
hako began on `jetpackio/devbox` (nix), then moved to plain Debian baking every
tool into the image via apt and `curl|bash`. That all-in-the-image model has two
faults: tools updated at runtime silently revert on the next rebuild (and we do
not want build-time auto-update), and the `curl|bash` installs are unpinned,
unverified remote code.

## Decision
Split into two tiers:
- **OS tier** — a thin, pinned Debian image: base, shells, sudo, tini, gmux,
  ffmpeg, and the `mise` binary. Built the Docker way; every download is version-
  and checksum-pinned.
- **Toolchain tier** — node, bun, python, pi, and the CLI tools, declared in
  `agent/.config/mise/config.toml` and pinned by `mise.lock`. mise installs them
  into the mounted home (`~/.local/share/mise`) at start (see `entrypoint.sh`).

## Why
Tools in the persistent home survive rebuilds, so runtime updates stick: pi can
be updated from inside, it persists, and the change shows up as a lock diff.
`mise.lock` gives reproducibility *and* integrity (versions + checksums; see
ADR-0008).

Not nix/devbox: their store lives in `/nix` (outside the home), so persisting it
needs a separate volume and fights the "everything in the home" model. mise is
home-native, ships a multi-platform lockfile, and carries our tools in its
registry (mostly checksum-backed aqua releases).

## Consequences
A cold clone installs the toolchain (~10-15s) in the background while gmux is
already up; later starts are fast (idempotent). Reproducibility now depends on
upstream releases still existing at install time (a baked image is more durable
here). New trust root: mise + the aqua registry. Media tools (ffmpeg) stay
OS-tier — heavy and awkward in mise.
