# hako

**An opinionated, sandboxed home for a coding agent.** Clone it, bring your own
agent credentials, and you get **pi** running in a container with live browser
access — without handing the agent your host or your keys. Customize by forking.

## Quickstart

```sh
git clone https://github.com/mgabor3141/hako && cd hako
docker compose up -d                  # build + start; gmux on :8791
# first start installs the dev toolchain into the home (~15s, in the background)
docker compose exec hako gmuxd auth   # prints a login URL + token
# open http://localhost:8791, authenticate, then:
docker compose exec hako gmux pi      # launch the agent (authenticate it once)
```

Your pi sessions show up live at <http://localhost:8791>. There's also a host
launcher — `./hako up`, `./hako shell`, `./hako pi`, … (run `./hako` for the list).

**Windows / WSL2:** run from inside your WSL2 distro (Docker Desktop WSL
integration on) and **clone into the Linux home (`~`), not `/mnt/c/...`** — bind
mounts and permissions only behave on the native filesystem.

## Architecture

```mermaid
flowchart LR
    you(["you (browser)"])
    fork[("your fork<br/>./agent")]
    up[("upstream MCP servers<br/>e.g. GitHub")]

    subgraph priv["Docker: hako private network"]
        subgraph agentc["agent container"]
            pi["pi agent"]
            gmuxd["gmuxd"]
            mise["dev toolchain (mise)"]
        end
        gw["MCP gateway<br/>holds creds + approval gate"]
    end

    you -->|"localhost:8791"| gmuxd
    fork -.->|"bind mount to /home/agent"| agentc
    pi -->|"CLI adapters (github, ...)"| gw
    gw -->|"credentials"| up
    gw -.->|"approval"| you
```

- **Your fork is the unit of customization.** `./agent/` is **bind-mounted as
  the agent's entire home**, so config, projects, and scratch all live in the
  repo. Nothing on your host (including `~/.pi`) is touched.
- **The agent holds no host credentials** — the security boundary is the
  *absence* of secrets, not behavior restrictions. When it needs a real tool it
  goes through the **MCP gateway**, which holds the creds and gates sensitive
  calls behind your approval *(phase 2; see Roadmap)*.

## Tools

hako leans on a few well-chosen tools so it stays small and legible:

- **Docker (Compose)** — the only thing you install on the host. Runs the
  sandbox and the private network between the agent and the gateway.
- **pi** — the coding agent you actually use; preconfigured with hako's opinions
  in `agent/.pi/agent/`. Bring your own provider; hako ships no credentials.
- **gmux** — browser access to terminal sessions: watch and attach to the agent
  live at `:8791` (loopback-only, token-authed).
- **mise** — pins and installs the in-home dev toolchain (node, bun, python,
  ripgrep, fd, jj, …) from a lockfile. It also enforces a **supply-chain release
  delay**: a new release isn't adopted until it has been public for ≥24h, so a
  freshly-compromised version can be caught or yanked before hako installs it.
- **bun** — runs pi and the MCP CLI adapters (one fast binary, no node_modules
  churn). `pi update` is rerouted through mise so the pinned core stays pinned.
- **MCP gateway** *(phase 2)* — a fork of [`mcp-proxy`](https://github.com/TBXark/mcp-proxy)
  that holds upstream credentials and exposes tools to the agent over the private
  network, gating chosen calls behind a swappable approval hook. The agent gets
  curated CLI **adapters** (e.g. `github`, in `agent/.agents/skills/`) and never
  sees the keys.

## Handy in the shell

The human shell (zsh) ships some niceties — run `help` inside the container for
the full, colorized list:

- `Ctrl-R` fuzzy history, `Ctrl-T` file picker, `Alt-C` fuzzy `cd` (fzf)
- `→` / `End` accepts the grey autosuggestion from your history
- `z <name>` jumps to a frequent dir (zoxide); a bare dir name `cd`s into it
- `ls`/`ll`/`la`/`lt`/`l.` are [eza](https://eza.rocks); `bat` and `man` are syntax-highlighted

pi itself always runs plain `/bin/bash` with stock `ls`/`grep`, so commands it
hands you paste-and-run unchanged.

## Customizing

hako is meant to be **forked and `git pull`ed** — opinions, including the pinned
`mise.lock`, surface as merge conflicts you resolve.

- **pi config:** `agent/.pi/agent/settings.json` — live (restart to reconcile).
- **toolchain:** add tools to `agent/.config/mise/config.toml`, then
  `mise install` (or restart) — live.
- **OS image:** `container/Dockerfile` — needs `docker compose up -d --build`.

The configuring agent's guide is [`AGENTS.md`](./AGENTS.md); design decisions
live in [`docs/`](./docs/).

## Backups (optional)

Off by default. To snapshot the agent home to a repo the agent can't reach,
follow [`docs/backups.md`](./docs/backups.md) (a restic sidecar) — or just hand
that file to your agent.

## Roadmap

- **Phase 1 — pi + container + gmux** *(done)*: clone-and-up opinionated pi.
- **Phase 2 — governed tools** *(in progress)*: the MCP gateway + CLI adapters +
  per-call approval, so the agent reaches real tools while holding no credentials.
