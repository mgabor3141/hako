# launcher/

The Go source for the `hako` host launcher (ADR-0012). Non-user-facing: people
run root `./hako`, the bootstrap. It builds this and caches the binary by a hash
of the source (rebuild only on change), then execs it:

- **host Go installed** -> builds directly (dev + power users).
- **no Go** -> builds in a **digest-pinned `golang` container** (Docker is hako's
  one hard dependency, so this always works), run as the host UID so the binary
  isn't root-owned. (`HAKO_FORCE_CONTAINER=1` forces this even with Go installed.)

Either way you run a build of the exact source in your checkout -- no downloaded
binary, nothing to pin or verify beyond the repo you cloned. Build directly with:

```sh
cd launcher && go build -o hako .
```

It reads the integration catalog (`integrations/*/integration.toml`) and the
user's `hako.toml`, **assembles** the stack (links enabled skills, merges the
gateway config, resolves settings into `.hako.env`, selects sidecars), and wraps
`docker compose` plus credential setup (`auth`) and vault unlock.

Status: **Phase A + A2 done** -- manifest-driven assembly, compose wrapping, and
an **in-process vault** (`filippo.io/age` + locked memory; a single multi-secret
`vault/secrets.age` under one global passphrase; the gateway no longer
decrypts). The shell launcher is retired. **Phase B** is done too: a `configure`
TUI (bubbletea) toggles integrations, sets typed settings, and sets up auth,
writing hako.toml. **Delivery** is build-from-source (host Go, else the pinned
container), cached by source hash -- no release pipeline, no downloaded binary,
so no skew between the binary and the source you pulled.
