# ADR-0014: Integrations are composable manifests selected by config

- **Status:** Accepted -- 2026-06-16. **Built and validated:** the catalog ships
  github (skill + gateway backend + secret), websearch (skill + sidecar +
  settings), and backups (sidecar + setting, no skill); `hako.toml` plus the
  `hako configure` TUI select the enabled set, and the assembler generates the
  skill links, the merged gateway config, the settings env, and the sidecar/
  secret wiring.

## Context
hako should accumulate many tools over its life (github, gcloud, Google
Workspace, jira, confluence, web search, a scraper, ...). Users need to **enable
the ones they want and disable the rest** without forking and without hand-
editing then gitignoring tracked files. Disabled tools should be **invisible to
the agent** -- absent from pi's context and unreachable through the gateway --
both to shrink attack surface and to keep the model's context lean as the
catalog grows. Some integrations also need a **sidecar container** (a self-hosted
MCP server, a scraper, a search engine).

## Decision
An **integration** is a self-describing unit in a shipped, in-repo catalog at
`integrations/<name>/`: an `integration.toml` (metadata + declared needs) plus
**any subset** of {a skill, a gateway-backend snippet, a sidecar compose
fragment, required secret(s), typed settings}. Skill-only, sidecar-only,
backend-only, and combinations are all valid.

Integrations declare **typed settings** (`string`/`number`/`bool`, with a default
and description) alongside their secrets; the user sets values per integration
in `hako.toml`. Resolved settings and secrets are exposed to assembly as env
vars (`HAKO_<INT>_<SETTING>` for settings; the declared `env` name for secrets),
so `gateway.json` and sidecar compose fragments stay declarative and just
reference `${...}`. A `[sidecar]` may name a bool setting via `enabled_by` to
gate whether its container spins up (so an integration can offer "use the
bundled sidecar" vs "point at my own URL" purely through settings).

The **enabled set** is a user-owned, gitignored `hako.toml` (on/off per
integration, plus the integration's declared settings). hako
ships a tracked `hako.example.toml`; a `hako configure` TUI manages the real
file. So choosing tools is **config, not a fork**: never a tracked-file edit,
never a manual gitignore, and `git pull` adds catalog entries without touching
`hako.toml` (no merge conflicts). This formalizes two customization tiers --
**config** (integrations) vs **fork** (deeper opinions, ADR-0001).

The running stack is **assembled from the enabled set**, so disabled means absent
at three layers:
- **skill** not linked into pi's skill dir -> pi cannot see it;
- **backend** not merged into the generated `gateway/config.json` -> the gateway
  cannot route it;
- **sidecar** not in the compose `-f` list -> the container does not exist.

**Live enable, within limits.** The whole catalog is mounted into the agent
container, so enabling a **skill-only** integration is just a symlink (written
from the host into the bind-mounted skill dir) that the next pi session picks up
-- no container restart. Integrations that add a **gateway backend** or a
**sidecar** need a stack reconcile (`hako up`): mcp-proxy does not hot-reload its
config, and a new sidecar needs compose; a gateway-config change re-seals the
vault (re-unlock). A future gateway config hot-reload could make those live too.

pi visibility is achieved by **hako owning pi's skill dir** (link enabled only),
not a pi extension; pi stays oblivious.

## Consequences
- `gateway/config.json`, pi's skill dir, the compose `-f` list, and the set of
  active vault secrets become **generated/selected artifacts** (gitignored), not
  hand-edited. The hand-written `config.json` retires.
- All secrets live in a **single vault under one global passphrase** (ADR-0011);
  one unlock decrypts everything the enabled set needs.
- Each sidecar is more private-network surface and another pinned image
  (ADR-0008). Credentials stay with the gateway/sidecar, never the agent -- the
  ADR-0002 boundary holds.
- **Supersedes ADR-0013's placement:** the inlined adapters no longer live
  permanently in the home; skills live in the catalog and are linked into pi's
  dir per the enabled set (github becomes the first manifest).
- The **assembler + `hako configure` TUI** (config merge, compose selection,
  masked secret prompts, vault wiring) is the feature that justifies the Go
  launcher (ADR-0012); it subsumes the clean secret-handling case. The Go
  launcher provides `configure`/`up`/`auth`/`unlock`/assemble; the earlier shell
  `./hako` is retired.
- **In-repo catalog only** for now: no third-party or user-dropped integrations,
  which defers a provenance/trust question.
