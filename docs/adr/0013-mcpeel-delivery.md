# ADR-0013: mcpeel delivered via skills.sh, pinned

- **Status:** Accepted — 2026-06-16. Not yet wired.

## Context
mcpeel is the agent-side CLI layer (ADR-0001) — the `github` CLI and friends
that talk to the broker/gateway. It is distributed as a **skills repo**: plain
TypeScript, no npm package, no build step, `git pull` to update (mcpeel's own
ADR-0005/0002). pi auto-discovers skills from `~/.pi/agent/skills/` and the
cross-tool `~/.agents/skills/` convention. We need it in the agent **as a
reference, not vendored** (ADR-0001) and **pinned** (ADR-0008).

## Decision
Deliver mcpeel with **`skills.sh`** (`vercel-labs/skills`, `npx skills add`) —
the ecosystem-standard skills installer, which has explicit **pi** support and
installs into `~/.agents/skills` (canonical copy + per-agent symlinks). Pin to
an **immutable commit SHA** in the source URL (and pin the tool itself,
`npx skills@<ver>`). Run it **inside the container** at entrypoint/setup — bun
and node are present, and `~/.agents/skills` is in the bind-mounted home, so it
persists across rebuilds like the mise toolchain.

mcpeel's `github` **replaces** the `pi-github` package (dropped from
`settings.json`). Because `skills.sh` installs the skill but not the bundled
CLI as a bare command, the `hako` setup/entrypoint adds the mcpeel-specific
step: `ln -s …/github.ts ~/.local/bin/github`. Auth is wired by env
(`MCP_GATEWAY_URL` → the broker; token optional per ADR-0007).

## Consequences
Standard, pi-aware delivery with `update`/`list`/`remove` for free; consistent
with "reference not vendor" and the runtime-install model. Two caveats:
- **Pin rigor is a git SHA, not a checksum** — content-addressed and immutable
  (reproducible), but a notch below `mise.lock`'s sha256 bar (ADR-0008).
- **skills.sh installs the skill, not the CLI-on-PATH** — hence the extra
  symlink step, which is mcpeel-specific and owned by hako's setup.

Rejected: **mise** (no npm package / skills backend; would force a build that
fights mcpeel's design) and a **git submodule** (exact SHA pin, but awkward
placement vs pi's skill-dir layout plus submodule UX).
