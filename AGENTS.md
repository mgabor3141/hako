# hako (for the configuring agent)

hako is an opinionated, sandboxed, ready-to-run home for a coding agent: **pi**
in a container, a thin Debian OS image plus a **mise**-managed dev toolchain in
the home, and **gmux** for browser access, shipped as a `git clone && docker
compose up`.

Canonical repo: <https://github.com/mgabor3141/hako>

## Two agents, two jobs

- **You** run on the user's host, in this repo. Your job is to **configure
  hako** — the image, the compose wiring, and the shipped pi config.
- The **in-container agent** (pi, at `/home/agent`) does the user's actual
  development work, sandboxed. It is not you; don't conflate the two configs.
  Its guidance lives in `agent/.pi/agent/AGENTS.md`, not here.

## Security boundary

The container holds **zero host credentials**. hako never touches the host's
`~/.pi`; the only thing shared in is the bind-mounted `agent/` of this repo.
Never commit secrets, tokens, or auth.

## Layout

- `container/Dockerfile` — the **thin OS tier**: Debian base, user `agent` (home
  `/home/agent`), shells, sudo, tini, and three pinned + checksum-verified
  binaries in `/usr/local/bin`: `mise`, `gmux`/`gmuxd`, `ffmpeg`. No dev
  toolchain here.
- `container/entrypoint.sh` — tini wrapper; reconciles the home toolchain in the
  background (`mise install`), then execs the CMD `gmuxd run` (foreground).
- `compose.yaml` — ports, mounts, env (below).
- `agent/` — the bind-mounted home. Ships the opinions: pi config
  (`.pi/agent/`), shell/git/starship config, and the **toolchain manifest +
  lock** (`.config/mise/config.toml`, `.config/mise/mise.lock`). Its `.gitignore`
  whitelists only those and ignores all installs, runtime state, and projects.
  Note: pi rewrites `.pi/agent/settings.json` at runtime (e.g.
  `lastChangelogVersion`); `git restore` it before committing config changes.

## Decisions

Load-bearing, reversible choices are recorded as lightweight ADRs in
[`docs/adr/`](docs/adr/). Read them before re-litigating why hako is Debian,
mounts the whole home, or holds no credentials. When you make or reverse such a
choice, add or supersede an ADR — don't bury the rationale in a commit alone.

## The rule that bites

All of `agent/` is bind-mounted over `/home/agent`, so **anything baked into the
home at build time is shadowed at runtime.** OS-tier tools go in `/usr` /
`/usr/local/bin`, never `/home/agent`. The dev toolchain is the deliberate
exception: mise *installs* it into the home at runtime (nothing baked to shadow),
so it persists across rebuilds and stays updatable. See ADR-0004/0005.

## Ports & mounts (in `compose.yaml`)

| Kind  | Value | Why |
|-------|-------|-----|
| mount | `./agent → /home/agent` | the agent's entire home: pi config + projects |
| port  | `127.0.0.1:8791 → 8790` | gmux web UI (loopback only, token-authed); host 8791 so a host gmux can coexist |
| env   | `GMUXD_LISTEN=0.0.0.0`  | gmuxd binds loopback by default; in a container it must bind all interfaces or the published port reaches nothing |

Add ports/mounts here as needed (e.g. a dev server the in-container agent
runs). Keep published ports on `127.0.0.1` unless the user explicitly wants
LAN/remote exposure (gmux's own remote path is its Tailscale listener).

## Updating & testing

- **Config + tool list** (`agent/.pi/agent/*`, `agent/.config/mise/config.toml`):
  live — edit, then `mise install` or restart to reconcile. No image rebuild.
- **Bump a tool**: `mise upgrade` (or edit + `mise lock`), commit the `mise.lock`
  diff.
- **OS image** (Dockerfile / mise / gmux / ffmpeg): `docker compose up -d --build`.
- Smoke test:
  `docker compose exec hako bash -lc 'node -v; pi --version; gmuxd status'`
- gmux login URL + token: `docker compose exec hako gmuxd auth`
