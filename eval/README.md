# eval — benchmarking hako's pi config on Terminal-Bench

This measures **hako's shipped pi configuration** — the agent home at
`agent/.pi/agent/` (AGENTS.md, settings, prompts, skills, and extensions such as
`goal.ts`) — on [Terminal-Bench 2.0](https://www.tbench.ai/), so config changes
can be judged on evidence rather than vibes.

It's a dev tool. The launcher and gateway never import anything here; a normal
hako user can ignore this folder entirely. The cost (Docker, API budget) is paid
only by whoever runs it.

## What it does (and doesn't)

We **keep** Terminal-Bench's per-task Docker environments and hidden tests — the
hard, valuable part — and **replace only the agent**: instead of TB's tmux
keystroke driver, we install pi + hako's config home into each task container and
run `pi -p "<instruction>"`. The hidden `tests/` then score the result. Harbor
(TB's official harness) does the deterministic build → run → test → record loop;
`adapter/hako_agent.py` is a thin `AbstractInstalledAgent` plugged into it.

The hako *container* is out of scope on purpose — every task brings its own
environment, so what actually moves the score is the config home, which travels
into each task container intact.

## Matrix

Two axes:

- **model** — any `provider/model` pi supports (key read from the provider's env var).
- **variant**:
  - `baseline` — `pi -p` with the config as shipped.
  - `goal` — same, plus `PI_GOAL_AUTOSTART=1` so `goal.ts` runs its
    "diligent user" loop headlessly (it normally needs `/goal`, which you can't
    type in print mode). This tests whether the goal loop raises completion.

The `baseline` vs `goal` **diff** is the point: it isolates one extension's
contribution, across models.

## Prerequisites

- Docker running.
- [Harbor](https://www.harborframework.com/docs/getting-started) installed in the
  active Python env (`uv tool install ...` / pip). Verify with an oracle run.
- Provider API keys exported (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, …).

## Run

```bash
# Smoke-test one task first (validates the headless goal loop end-to-end).
TASKS=hello-world MODELS=anthropic/claude-sonnet-4-5 ./run.sh

# Full sweep
MODELS="anthropic/claude-sonnet-4-5 openai/gpt-5" VARIANTS="baseline goal" ./run.sh
```

Results land in `results/<timestamp>/<model>.<variant>/` (gitignored).

## Reproducibility (ADR-0008)

`run.sh` pins the inputs that determine a score: `TB_DATASET` (task set + version),
`PI_VERSION`, and `HAKO_REF` (the config commit under test). Pin `PI_VERSION` to
an exact version for any run you intend to cite.

## Anchoring PR deltas

Raw results are gitignored, but a PR that changes the config should cite numbers.
Commit a short summary (scores + the pins they came from) to `BASELINE.md` for
reference runs; keep transcripts local.

## Runtime notes

- **Headless goal loop — validated.** `pi -p` (and `--mode json`) do process the
  follow-up messages `goal.ts` injects on `agent_end`: a local smoke run with
  `PI_GOAL_AUTOSTART=1` autostarted the loop without `/goal` and iterated
  (progressive nudges, agent acted on each). No `--mode rpc` fallback needed.
- **The goal cell is slow and can run long.** The loop may approach
  `MAX_PASSES` (10) before the proxy emits `STOP`, especially with slower/terser
  models (the proxy takes the agent at its word and only sees prose, not tool
  results). Budget generous per-task timeouts for the `goal` variant; watch for
  cells that hit the task timeout rather than completing.
