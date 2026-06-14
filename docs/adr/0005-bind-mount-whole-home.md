# ADR-0005: Bind-mount the whole home; bake tools outside it

- **Status:** Accepted — 2026-06-14 (supersedes the earlier "chezmoi-apply this repo" idea)

## Context
How does hako deliver its config and persist agent state, hermetically?

## Decision
Bind-mount the repo's `agent/` as the agent's entire `/home/agent`. Ship config
as plain files. Bake all tooling **outside** the home (`/opt`, `/usr`). pi uses
its default `~/.pi/agent` — no env override.

## Why
Simplest hermetic model: ships as a `git clone`, updates via `git pull` (opinions
surface as merge conflicts, not silent clobbers), and never touches the host's
`~/.pi`. The mount shadows anything at `/home/agent`, so tools must live outside
it — which also forces ADR-0004. Rejected chezmoi-apply: it invents a config
format and machinery we don't need.

## Consequences
The whole home is the user's (config + projects + scratch). Runtime state
persists in the clone, gitignored. Caveat: the bind mount assumes host uid 1000
== container `agent`; mismatches need handling.
