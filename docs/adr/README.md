# Architecture decisions

Lightweight ADRs, one decision per file. Newer ADRs may supersede older ones;
mark the change in both rather than rewriting history.

- [0001](0001-three-repos.md) — Three repos, not a monorepo
- [0002](0002-no-host-credentials.md) — The agent holds no host credentials
- [0003](0003-auth-is-pis-job.md) — Auth is pi's job; hako is auth-agnostic
- [0004](0004-debian-not-devbox.md) — Debian base, not nix/devbox
- [0005](0005-bind-mount-whole-home.md) — Bind-mount the whole home; bake tools outside it
- [0006](0006-gmux-wiring.md) — gmux container wiring (bind address, port, token)
- [0007](0007-host-side-broker.md) — Host-side MCP broker (proposed)
