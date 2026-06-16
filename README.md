# hako

an opinionated, sandboxed coding agent config

Clone it, bring your own agent credentials, and you get **pi** running in a
container with browser access via **gmux** — without handing the agent your
host or your keys. Customize by forking.

## Quickstart

```sh
git clone https://github.com/mgabor3141/hako && cd hako
docker compose up -d                  # builds the image, starts gmux on :8791
# first start installs the dev toolchain into the home in the background (~15s)
docker compose exec hako gmuxd auth   # prints a login URL + token
# open http://localhost:8791, authenticate, then:
docker compose exec hako gmux pi      # launch the agent (authenticate it once)
```

Your pi sessions show up live at <http://localhost:8791>.

**Windows / WSL2:** run from inside your WSL2 distro (Docker Desktop WSL
integration enabled), and **clone into the Linux home (`~`), not `/mnt/c/...`**
— bind mounts and permissions only behave on the native filesystem.

## What's inside

- **pi** — the coding agent, preconfigured with hako's opinions
  (`agent/.pi/agent/`). Bring your own provider; hako ships no credentials.
- **gmux** — see and attach to every session from your browser (`:8791`,
  localhost-only, token-authed).
- A dev toolchain (node, bun, python, ripgrep, fd, jj, …) managed by **mise**
  and pinned by a lockfile. A thin, pinned OS image carries only the base,
  gmux, ffmpeg, and mise; the toolchain installs into the home on first start.

## Handy in the shell

The human shell (zsh) ships some niceties worth knowing. Run `help` inside the
container for the full, colorized list:

- `Ctrl-R` fuzzy history search, `Ctrl-T` file picker, `Alt-C` fuzzy `cd` (fzf)
- `→` / `End` accepts the grey autosuggestion from your history
- `z <name>` jumps to a frequent directory (zoxide); typing a bare dir name `cd`s into it
- `ls`/`ll`/`la`/`lt`/`l.` are [eza](https://eza.rocks); `bat` and `man` are syntax-highlighted

pi itself always runs plain `/bin/bash` with stock `ls`/`grep`, so commands it
hands you paste-and-run unchanged.

## How it's wired

- `agent/` is bind-mounted as the agent's entire home — config, projects,
  scratch. Nothing on your host (including `~/.pi`) is touched.
- The agent holds **no host credentials**: the boundary is the absence of
  secrets, not behavior restrictions.
- Config and the tool list (`agent/.config/mise/config.toml`) are live: edit and
  restart to reconcile. Only OS-image changes need `docker compose up -d --build`.

## Customizing

Edit `agent/.pi/agent/settings.json` (pi's config) or add tools to
`agent/.config/mise/config.toml` then `mise install` (or restart) — both live.
The OS image (`container/Dockerfile`) needs a rebuild. hako is meant to be forked
and `git pull`ed — opinions, including the pinned `mise.lock`, surface as merge
conflicts. The configuring agent's guide is
[`AGENTS.md`](./AGENTS.md); design decisions are in [`docs/`](./docs/).

## Backups (optional)

Not on by default. To snapshot the agent home to a repo the agent can't reach,
follow [`docs/backups.md`](./docs/backups.md) (a restic sidecar) — or just hand
that file to your agent.

## Roadmap

- **Phase 1 — pi + container + gmux** *(current)*: clone-and-up opinionated pi.
- **Phase 2 — governed tools**: a pinned MCP gateway + skills, so agents reach
  tools through a gateway that holds the credentials and the agent holds none.
