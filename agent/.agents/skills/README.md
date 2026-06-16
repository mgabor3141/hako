# hako MCP CLI adapters

Small, agent-optimized CLIs that talk to an MCP endpoint (the hako gateway, or
any MCP gateway). Each lives in its own skill folder (`<name>/SKILL.md` + a
TypeScript CLI). They're curated for agent ergonomics: minimal token cost, one
call per question, errors that teach the next action. Formerly the standalone
`mcpeel` project, now inlined here — hako is the reference implementation.

## Running them

- **Runtime:** `bun` (preferred) or `node` >= 22.18 — both ship in hako's
  toolchain.
- **In hako:** the CLIs are placed on `PATH` for you (e.g. `github`) and the
  endpoint is wired via `MCP_GATEWAY_URL` (gateway token optional). Just run
  `github --help`.
- **Standalone (outside hako):** make the CLI invocable
  (`ln -s <skill>/github.ts ~/.local/bin/github`, or run `bun <path>`), and set
  `MCP_GATEWAY_URL` (+ `MCP_GATEWAY_TOKEN` only if your gateway requires one),
  or the per-tool `GITHUB_MCP_URL` / `GITHUB_MCP_TOKEN`.

## Auth boundary

The CLI never holds upstream credentials — it talks to a gateway that does (the
hako gateway). Never hardcode tokens in a `SKILL.md` or any committed file.

## Editing and contributing

These are opinionated and meant to be edited directly (that is the
customization mechanism). Each folder ships contract tests that run against a
local mock — no credentials needed:

```sh
cd github && bun test.ts   # or: node test.ts
```

hako is fork-and-`git pull`, so your edits surface as merge conflicts rather
than silent clobbers.
