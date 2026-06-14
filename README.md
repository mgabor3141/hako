# hako

**箱** — an opinionated, sandboxed agent harness you run in one command.
*hako* is Japanese for "box"; *hakoniwa* (箱庭, "boxed garden") is the Japanese
word for a sandbox — which is exactly what this is.

hako is to [pi](https://github.com/) what a curated config is to a bare editor:
a strong set of opinions on top of pi, **plus** a containerized environment,
**plus** [gmux](https://github.com/gmuxapp) for connectivity — and, later, a
governed MCP layer. Clone it, bring your own pi auth, `docker compose up`, and
you have a working agent. Customize by forking.

## What's in the box

| Layer | What it is |
|---|---|
| **Opinions** | A curated pi configuration (settings, skills, prompts) |
| **Environment** | A reproducible container (devbox toolchain + pi + gmux) |
| **Connectivity** | gmux — attach to live agent sessions from a browser |
| **Governance** *(later)* | MCP broker + mcpeel skills, so agents reach tools without holding upstream credentials |

hako **integrates**, it doesn't absorb: the container image lives here, but the
broker and the mcpeel skills are referenced/pinned, not vendored — they stay
reusable on their own.

## Authentication

hako ships **no credentials and assumes no provider.** Auth is entirely pi's
job — bring your own (any provider). You authenticate pi *once* inside the
container; it persists in the home volume across restarts. hako never sees a
key.

## Quickstart

```sh
git clone https://github.com/<you>/hako && cd hako
docker compose up -d           # builds the image, starts the agent
docker compose exec hako bash  # shell in
# one-time, inside the container:
pi auth                        # authenticate pi with your provider of choice
gmux pi                        # start an agent session (attach via browser)
```

State (pi auth, gmux sessions, your work) lives in a named volume, so restarts
don't lose it.

## Roadmap

- **Phase 1 — pi + container + gmux** *(current)*: clone-and-up opinionated pi.
- **Phase 2 — governed tools**: pin the MCP broker + mcpeel skills; agents call
  tools through a broker that holds the credentials, so the agent environment
  holds none.

## Repository layout

```
container/      the agent image (Dockerfile, entrypoint) — canonical home
config/         the opinionated chezmoi-applied config (pi settings, devbox.json, shell glue)
compose.yaml    the deployment: one agent service, persistent home volume
docs/           architecture decisions
```

## How customization works

The container bootstraps by applying *this repo* as its config
(`chezmoi init --apply <this-repo>`). To make it yours, **fork hako and point
the bootstrap at your fork** — edit the pi settings, the package list, whatever.
No credentials live in the repo, so forks are safe to make public.

## Design notes

See [`docs/`](./docs/) for the architecture decisions (the three-repo split, the
agent-holds-no-credentials boundary, why the broker is separate, etc.).
