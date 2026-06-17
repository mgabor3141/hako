# launcher/

The Go source for the `hako` host launcher (ADR-0012). Non-user-facing: people
run root `./hako`, the bootstrap. It has two paths:

- **Go installed** -> builds this from source and execs it (dev + power users).
- **no Go** -> downloads the pinned release binary for the host and verifies its
  sha256 against the committed `launcher/checksums.txt` before running it. No
  toolchain needed. (`HAKO_DOWNLOAD=1` forces this path even with Go installed.)

Build directly with:

```sh
cd launcher && go build -o hako .
```

It reads the integration catalog (`integrations/*/integration.toml`) and the
user's `hako.toml`, **assembles** the stack (links enabled skills, merges the
gateway config, resolves settings into `.hako.env`, selects sidecars), and wraps
`docker compose` plus the vault seal/unseal.

Status: **Phase A + A2 done** -- manifest-driven assembly, compose wrapping, and
an **in-process vault** (`filippo.io/age` + locked memory; a single multi-secret
`vault/secrets.age` under one global passphrase; the gateway no longer
decrypts). The shell launcher is retired. **Phase B** is done too: a `configure`
TUI (bubbletea) toggles integrations, sets typed settings, and seals secrets,
writing hako.toml. **Phase C** is in place: `.goreleaser.yaml` +
`.github/workflows/release.yml` build pinned, checksummed binaries on a `v*` tag,
and the bootstrap downloads + verifies them.

## Cutting a release

```sh
git tag v0.1.0 && git push origin v0.1.0   # CI runs goreleaser -> GitHub release
```

Then **pin it** so the bootstrap (and `git pull` users) adopt it: download the
release's `checksums.txt` to `launcher/checksums.txt` and write the tag to
`launcher/HAKO_VERSION`, and commit both. Pinning the checksums in-repo (not
trusting the release blob) is the supply-chain point (ADR-0008): a tampered
release can't change them without a diff you see on pull.
