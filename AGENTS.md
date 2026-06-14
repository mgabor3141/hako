# hako (for the configuring agent)

hako is an opinionated, sandboxed, ready-to-run home for a coding agent: **pi**
in a container, a Debian dev box, and **gmux** for browser access, shipped as a
`git clone && docker compose up`.

## Two agents, two jobs

- **You** run on the user's host, in this repo. Your job is to **configure
  hako** — the image, the compose wiring, and the shipped pi config.
- The **in-container agent** (pi, at `/home/agent`) does the user's actual
  development work, sandboxed. It is not you; don't conflate the two configs.
  Its guidance lives in `home/.pi/agent/AGENTS.md`, not here.

## Security boundary

The container holds **zero host credentials**. hako never touches the host's
`~/.pi`; the only thing shared in is the bind-mounted `home/` of this repo.
Never commit secrets, tokens, or auth.

## Layout

- `container/Dockerfile` — Debian base; user `agent` (home `/home/agent`); all
  tooling baked **outside the home**: apt CLI tools + node (NodeSource) in
  `/usr`, bun + pi in `/opt/bun`, gmux in `/usr/local/bin`.
- `container/entrypoint.sh` — tini wrapper; the CMD is `gmuxd run` (foreground).
- `compose.yaml` — ports, mounts, env (below).
- `home/` — the bind-mounted home. Ships pi's opinions
  (`.pi/agent/settings.json`, `.pi/agent/AGENTS.md`); its `.gitignore`
  whitelists only those and ignores all runtime state and the user's projects.

## Decisions

Load-bearing, reversible choices are recorded as lightweight ADRs in
[`docs/adr/`](docs/adr/). Read them before re-litigating why hako is Debian,
mounts the whole home, or holds no credentials. When you make or reverse such a
choice, add or supersede an ADR — don't bury the rationale in a commit alone.

## The rule that bites

All of `home/` is bind-mounted over `/home/agent`, so **anything installed into
the home at build time is shadowed at runtime.** Bake tools into `/opt` or
`/usr`, never `/home/agent`. (This is why hako is Debian, not nix/devbox — nix
lives in the home and fights the mount.)

## Ports & mounts (in `compose.yaml`)

| Kind  | Value | Why |
|-------|-------|-----|
| mount | `./home → /home/agent` | the agent's entire home: pi config + projects |
| port  | `127.0.0.1:8790 → 8790` | gmux web UI (loopback only, token-authed) |
| env   | `GMUXD_LISTEN=0.0.0.0` | gmuxd binds loopback by default; in a container it must bind all interfaces or the published port reaches nothing |

Add ports/mounts here as needed (e.g. a dev server the in-container agent
runs). Keep published ports on `127.0.0.1` unless the user explicitly wants
LAN/remote exposure (gmux's own remote path is its Tailscale listener).

## Updating & testing

- **Config** (`home/.pi/agent/*`): live — edit, no rebuild.
- **Image** (Dockerfile / tools / pi / gmux): `docker compose up -d --build`.
- Smoke test:
  `docker compose exec hako bash -lc 'node -v; pi --version; gmuxd status'`
- gmux login URL + token: `docker compose exec hako gmuxd auth`
