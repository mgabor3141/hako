# hako

**An opinionated, sandboxed home for a coding agent.** Clone it, bring your own
agent credentials, and you get **pi** running in a container you watch and steer
from your browser -- without handing the agent your host or your keys. Fork it to
make it yours.

## Quickstart

```sh
git clone https://github.com/mgabor3141/hako && cd hako
./hako up            # build + start; the first run installs the toolchain (~15s)
./hako token         # print a login URL + token for the browser UI
# open http://localhost:8791, sign in, then:
./hako pi            # launch the agent (authenticate your provider once)
```

Your pi sessions show up live at <http://localhost:8791>. `./hako` is a tiny
bootstrap that builds a small Go binary (with your Go, or in a pinned container
-- Docker is enough), then assembles your enabled integrations and drives
`docker compose`. Run `./hako` on its own for the full command list.

**Windows / WSL2:** run from inside your WSL2 distro (Docker Desktop WSL
integration on) and **clone into the Linux home (`~`), not `/mnt/c/...`** -- bind
mounts and permissions only behave on the native filesystem.

## How it works

```mermaid
flowchart LR
    you(["you (browser)"])
    fork[("your fork:<br/>agent/ + hako.toml")]
    up[("upstream MCP servers<br/>e.g. GitHub")]
    vault[("age vault")]

    subgraph priv["Docker: hako private network"]
        subgraph agentc["agent container"]
            pi["pi + enabled skills"]
            gmuxd["gmuxd"]
        end
        gw["MCP gateway<br/>holds creds + approval gate"]
        sc["sidecars<br/>e.g. web search"]
    end

    you -->|"localhost:8791"| gmuxd
    fork -.->|"bind mount + assemble"| agentc
    pi -->|"CLI adapters"| gw
    pi -->|"direct"| sc
    gw -->|"credentials"| up
    vault -.->|"hako unlock (host)"| gw
    gw -.->|"approve?"| gmuxd
```

- **Your fork is the unit of customization.** `./agent/` is bind-mounted as the
  agent's entire home -- config, projects, and scratch all live in the repo.
  Nothing on your host (including `~/.pi`) is touched.
- **The agent holds no credentials.** The boundary is the *absence* of secrets,
  not behavior rules. For a real tool it goes through the **MCP gateway**, which
  holds the credentials and gates sensitive calls behind your approval -- a `y/N`
  that appears as a session in the gmux dashboard. Which tools exist is up to the
  **integrations** you enable.
- **Reproducible and pinned.** The in-home toolchain is locked (`mise.lock`) and
  the gateway image is digest-pinned; mise also waits ~24h before adopting a new
  release, so a freshly compromised version can be caught or yanked before hako
  installs it.

## Integrations

What the agent can reach is composable. Each lives in `integrations/` -- a skill
the agent calls, plus whatever it needs (a gateway backend, a sidecar, secrets)
-- and you toggle them in **`hako.toml`** (gitignored, so choosing tools is never
a merge conflict). Disabled ones are invisible to the agent: no skill, no gateway
route, no sidecar.

```sh
./hako configure     # TUI: toggle integrations, set options, set up auth
./hako up            # assemble + start only what's enabled
```

Shipped today: **github** (PRs/issues/CI through the gateway), **websearch** (a
SearXNG sidecar), and **webview** (read a page as markdown via crawl4ai; off by
default -- heavy image). Add your own by dropping a folder in `integrations/` --
see [`integrations/README.md`](./integrations/README.md).

## Credentials

The agent holds none. Tokens live in a single **age-encrypted vault** under one
passphrase, decrypted on your host at unlock time and handed straight to the
gateway -- never written to disk.

```sh
./hako auth github   # guided: which token + scopes to create; sets a passphrase
./hako up            # prompts for the passphrase and unlocks
./hako unlock        # re-enter it after a gateway restart
```

`hako auth <integration>` walks you through exactly what to create and which
scopes to grant.

## Customizing

hako is meant to be **forked and `git pull`ed** -- opinions (including the pinned
`mise.lock`) surface as merge conflicts you resolve.

- **agent + pi config:** edit under `agent/` (e.g. `agent/.pi/agent/settings.json`);
  live on restart.
- **toolchain:** add tools to `agent/.config/mise/config.toml`, then `mise install`.
- **OS image:** `container/Dockerfile` -- rebuild with `./hako up --build`.

The guide for an agent customizing hako is [`AGENTS.md`](./AGENTS.md); design
decisions live in [`docs/`](./docs/).

## More

- **Shell niceties** -- the human shell (zsh) has fzf history/file-picker,
  autosuggestions, zoxide, and eza/bat; run `help` inside the container. pi itself
  always runs plain bash with stock tools, so commands it hands you paste-and-run
  unchanged.
- **Let the agent push** -- set up a scoped, transport-only git credential (the
  one deliberate exception to zero-cred): [`docs/git.md`](./docs/git.md).
- **Drive it from your editor** over [ACP](https://agentclientprotocol.com) (e.g.
  Zed): [`docs/acp.md`](./docs/acp.md).
- **Back up** the agent home to a repo it can't reach: [`docs/backups.md`](./docs/backups.md).
- **Not yet hardened for real-credential use** --
  [`docs/production-readiness.md`](./docs/production-readiness.md) is the honest
  list of rough edges.
