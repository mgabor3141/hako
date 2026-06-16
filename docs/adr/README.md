# Architecture decisions

Lightweight ADRs, one decision per file. The project is still prototyping, so an
ADR may be edited in place when a decision changes (rather than kept as a
supersession chain); the status line notes what it supersedes.

- [0001](0001-three-repos.md) — Three repos, not a monorepo
- [0002](0002-no-host-credentials.md) — The agent holds no host credentials
- [0003](0003-auth-is-pis-job.md) — Auth is pi's job; hako is auth-agnostic
- [0004](0004-debian-not-devbox.md) — Two tiers: thin Debian OS image + mise-managed home toolchain
- [0005](0005-bind-mount-whole-home.md) — Bind-mount the whole home; OS tools outside, toolchain inside
- [0006](0006-gmux-wiring.md) — gmux container wiring (bind address, port, token)
- [0007](0007-host-side-broker.md) — The broker is a remote-proxy container sidecar
- [0008](0008-pinning-and-integrity.md) — Reproducibility and supply-chain integrity by pinning
- [0009](0009-backups-optional-addon.md) — Backups are an opt-in documented add-on
- [0010](0010-tool-call-approval.md) — Tool-call approval is a swappable hook
- [0011](0011-secrets-passphrase-vault.md) — Secrets: passphrase-vault by default
- [0012](0012-host-launcher.md) — Host-side `hako` launcher
