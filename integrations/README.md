# hako integrations

The catalog of things hako can wire up: GitHub, and (over time) gcloud, Google
Workspace, Jira, Confluence, web search, and more. Each lives in its own folder
and is a **self-describing manifest** plus whatever it provides.

You enable the ones you want in `hako.toml` (gitignored, managed by
`hako configure`); disabled integrations are invisible to the agent. See
[ADR-0014](../docs/adr/0014-integrations-as-manifests.md).

## Anatomy of an integration

```
integrations/<name>/
  integration.toml   # name, summary, declared skill/gateway/sidecar/secrets/settings
  skill/             # OPTIONAL: a CLI adapter linked into pi's skill dir
  gateway.json       # OPTIONAL: an mcpServers entry merged into the gateway config
  compose.yaml       # OPTIONAL: a sidecar container
```

Any subset is valid: skill-only, sidecar-only, backend-only, or combinations.
At `hako up` the enabled set is assembled into the running stack (skills linked,
gateway config generated, sidecars composed, secrets pulled from the vault).

## The skill adapters

Skills are small, agent-optimized CLIs that talk to an MCP endpoint (the hako
gateway). Curated for agent ergonomics: minimal token cost, one call per
question, errors that teach the next action. They run under `bun` (in hako the
CLI is placed on `PATH` and its gateway endpoint is wired for you). The CLI never
holds upstream credentials -- the gateway does. Each ships contract tests
against a local mock (no credentials):

```sh
cd <name>/skill && bun test.ts
```

These are opinionated and meant to be edited directly; hako is fork-and-`git
pull`, so your edits surface as merge conflicts rather than silent clobbers.

Not every skill is a gateway adapter. An **instructional** skill ships only a
`SKILL.md` (no CLI) and teaches the agent to use what's already in the
container -- e.g. `office`, which has pi write short Python against bundled
libraries (python-docx / openpyxl / python-pptx).
