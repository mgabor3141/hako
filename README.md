# hako

**箱** — an opinionated, sandboxed home for a coding agent, in one command.
*hako* is Japanese for "box"; *hakoniwa* (箱庭) means sandbox.

Clone it, bring your own agent credentials, and you get **pi** running in a
container with browser access via **gmux** — without handing the agent your
host or your keys. Customize by forking.

## Quickstart

```sh
git clone https://github.com/mgabor3141/hako && cd hako
docker compose up -d                  # builds the image, starts gmux on :8790
docker compose exec hako gmuxd auth   # prints a login URL + token
# open http://localhost:8790, authenticate, then:
docker compose exec hako gmux pi      # launch the agent (authenticate it once)
```

Your pi sessions show up live at <http://localhost:8790>.

**Windows / WSL2:** run from inside your WSL2 distro (Docker Desktop WSL
integration enabled), and **clone into the Linux home (`~`), not `/mnt/c/...`**
— bind mounts and permissions only behave on the native filesystem.

## What's inside

- **pi** — the coding agent, preconfigured with hako's opinions
  (`agent/.pi/agent/`). Bring your own provider; hako ships no credentials.
- **gmux** — see and attach to every session from your browser (`:8790`,
  localhost-only, token-authed).
- A Debian dev box (git, ripgrep, fd, bun, node, …) baked outside the agent's
  home, so the whole home (`agent/`) is yours.

## How it's wired

- `agent/` is bind-mounted as the agent's entire home — config, projects,
  scratch. Nothing on your host (including `~/.pi`) is touched.
- The agent holds **no host credentials**: the boundary is the absence of
  secrets, not behavior restrictions.
- Config is live; image changes need `docker compose up -d --build`.

## Customizing

Edit `agent/.pi/agent/settings.json` (live) or `container/Dockerfile` (rebuild).
hako is meant to be forked and `git pull`ed — opinions surface as merge
conflicts, not silent clobbers. The configuring agent's guide is
[`AGENTS.md`](./AGENTS.md); design decisions are in [`docs/`](./docs/).

## Roadmap

- **Phase 1 — pi + container + gmux** *(current)*: clone-and-up opinionated pi.
- **Phase 2 — governed tools**: a pinned MCP broker + skills, so agents reach
  tools through a broker that holds the credentials and the agent holds none.
