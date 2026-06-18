# ADR-0008: Reproducibility and supply-chain integrity by pinning

- **Status:** Accepted — 2026-06-15

## Context
hako ships as a clone-and-build to people who are not the author. Two reasons to
care how deterministic the build is. **Security:** every unpinned `curl|bash` or
`latest` is fresh third-party code, and for an autonomous-agent box whose whole
boundary is "no credentials" (ADR-0002), a poisoned tool can act without needing
any. **Reliability:** a build that worked should work again, and failures should
correlate with deliberate bumps, not the calendar.

## Decision
Pin and verify everything, per tier:
- **OS tier**: pinned base image; mise, gmux, and ffmpeg downloaded by version
  and verified by `sha256`. No `curl|bash` installers.
- **Toolchain tier**: `mise.lock` pins exact versions and checksums for every
  tool across platforms; installs run `--locked`.
- **Release-age delay**: mise's `minimum_release_age` (default 24h) stays on, so
  a freshly-published version isn't adopted until it has aged enough for a
  compromise to be caught or yanked (it skipped a <24h-old release in practice).
- **Gateway tier**: the MCP gateway image is **pinned by `@sha256` digest** to a
  specific fork-commit CI build (`ghcr.io/mgabor3141/mcp-proxy`, published per
  commit by the fork's `image` workflow). No semver releases -- commit-hash
  identity, immutable digest. Bumping is a one-line digest diff in
  `gateway/compose.gateway.yaml`.

Bumps are deliberate, reviewable commits (a hash or lock diff).

## Why
Version strings are not integrity; checksums are. A real lockfile also pins the
transitive closure (the xz lesson). Reliability matters less here than for prod —
the config in the home is unversioned and a user plus an agent are present to fix
breakage — so the goal is a trustworthy authoritative tier and a clean handoff
baseline, not bit-perfect determinism.

## Consequences
"Scratch" is allowed and expected: the agent or user can install at runtime to
fix or explore. Those changes are ephemeral (revert on rebuild) or, in the home,
visible as diffs, and you "promote" them by editing the manifest/lock. Pinning
trades auto-patching for vetted integrity, so it needs a bump cadence (stale pins
carry known CVEs). gmux shares the author's trust root, so pinning it defends
only the *independent*-compromise case; hako users already extend the author full
trust by running the build, but a security-conscious forker should treat gmux as
a coupled dependency.
