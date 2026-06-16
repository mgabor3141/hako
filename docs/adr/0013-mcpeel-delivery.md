# ADR-0013: The MCP CLI adapters are inlined into hako

- **Status:** Accepted — 2026-06-16 (supersedes the same-date "deliver via
  skills.sh, pinned to a SHA" decision). Not yet wired.

## Context
The agent-side MCP CLI adapters (the `github` CLI etc., formerly the standalone
"mcpeel") are a skills repo: plain TypeScript, no npm package, no build step.
pi auto-discovers skills from `~/.pi/agent/skills/` and the cross-tool
`~/.agents/skills/` convention. We first planned to fetch them via `skills.sh`
(`npx skills add`) pinned to a commit SHA — but the adapters are co-designed and
co-tested with hako, the cross-repo SHA pin sat below our checksum bar
(ADR-0008), and hako *is* the integration layer whose fork-and-pull model
(ADR-0005) already fits "edit the source, keep a fork."

## Decision
**Inline the adapters into hako**, in-tree under `agent/.agents/skills/<name>/`
(where pi reads them via the bind-mounted home), with a **root-level `skills`
symlink → `agent/.agents/skills`** so they're also exposed at the conventional
discovery path. hako **owns the PATH/symlink setup** for the CLIs (e.g.
`github` → `~/.local/bin`), done in the launcher/entrypoint.

The standalone **`mcpeel` meta-skill is dropped** — its only job was teaching
self-install/management, which hako now handles. A **README in the skills dir**
carries the relevant runtime/auth notes (TS runtime, `MCP_GATEWAY_URL`, how to
run a CLI standalone). hako is the **reference implementation**: anyone or any
agent can still pull or `skills.sh`-install the adapters from here.

The `github` adapter **replaces** the `pi-github` package (dropped from
`settings.json`). Auth is wired by env (`MCP_GATEWAY_URL` → the broker; token
optional per ADR-0007).

## Consequences
In-tree means **reproducible by construction** — no cross-repo pin, no fetch,
no "is the pinned SHA the tested SHA?" gap. The adapters ride hako's CI (which
gains a TS bun+node job) and propagate as fork-and-pull merge conflicts.
`skills.sh` is demoted from hako's delivery mechanism to the external-consumption
convenience. Costs: hako's repo gains TS and the adapters' tests; the standalone
"mcpeel" name goes away (one fewer coined name to learn). Follow-ups: whitelist
the skill files in `agent/.gitignore`, decide the fate of the old mcpeel repo
(archive/redirect).
