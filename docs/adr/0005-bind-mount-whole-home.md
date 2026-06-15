# ADR-0005: Bind-mount the whole home; OS tools outside it, toolchain inside

- **Status:** Accepted — 2026-06-15 (updated for the two-tier model of ADR-0004; supersedes the earlier "chezmoi-apply this repo" idea)

## Context
How does hako deliver config, persist agent state, and carry the dev toolchain,
hermetically?

## Decision
Bind-mount the repo's `agent/` as the agent's entire `/home/agent`. Ship config
as plain files. **OS-tier** tooling is baked outside the home (`/usr`,
`/usr/local/bin`); the **dev toolchain** lives inside the home, installed by mise
(ADR-0004). pi uses its default `~/.pi/agent` — no env override.

## Why
Simplest hermetic model: ships as a `git clone`, updates via `git pull` (opinions
surface as merge conflicts, not silent clobbers), and never touches the host's
`~/.pi`. The mount shadows anything baked at `/home/agent`, so **OS** tools must
live outside it. The dev toolchain is the deliberate exception: it is *installed*
into the home at runtime (nothing baked to shadow), so it persists and stays
updatable.

## Consequences
The whole home is the user's (config + toolchain + projects + scratch). Tool
installs and runtime state persist in the clone, gitignored; only opinions
(config plus the mise manifest and lock) are tracked. Caveats: the mount assumes
host uid 1000 == container `agent`, and the toolchain dir is vulnerable to
`git clean -fdx` in the clone.
