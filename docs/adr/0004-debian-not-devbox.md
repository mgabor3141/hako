# ADR-0004: Debian base, not nix/devbox

- **Status:** Accepted — 2026-06-14 (supersedes the initial jetpackio/devbox base)

## Context
hako started on the `jetpackio/devbox` image (nix-based, declarative packages
via `devbox.json`), inherited from a personal dev container.

## Decision
Use a plain **Debian** base with apt for CLI tools and bun for pi — not
nix/devbox.

## Why
nix stores its per-user profile, channels, and state **inside the home**
(`~/.local/state/nix`, `~/.nix-*`). Bind-mounting the whole home (ADR-0005)
shadows all of that and breaks nix. Debian keeps the home empty of machinery.

## Consequences
Simpler, faster builds and a clean whole-home mount. We lose nixpkgs breadth and
the declarative `devbox.json`; acceptable, since hako's package set is small and
apt-pinned is reproducible enough.
