# launcher/

The Go source for the `hako` host launcher (ADR-0012). Non-user-facing: people
run root `./hako`, which is a bootstrap that builds this and execs it (and will
fetch a pinned release binary once CI publishes them). Build directly with:

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
decrypts). The shell launcher is retired. Next: **Phase B** (the `configure`
TUI) and **Phase C** (CI release binaries the bootstrap downloads + verifies).
