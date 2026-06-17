# launcher/

The Go source for the `hako` host launcher (ADR-0012). Non-user-facing: people
run the compiled `hako` binary (eventually delivered by a pinned bootstrap), not
this source.

```sh
cd launcher && go build -o hako .   # the binary is runnable from the repo root
```

It reads the integration catalog (`integrations/*/integration.toml`) and the
user's `hako.toml`, **assembles** the stack (links enabled skills, merges the
gateway config, resolves settings into `.hako.env`, selects sidecars), and wraps
`docker compose` plus the vault seal/unseal.

Status: **Phase A** -- parity with the shell `./hako` (which stays as the
stopgap), but manifest-driven. Next: **A2** moves the vault in-process
(`filippo.io/age` + locked memory, a single multi-secret vault); **Phase B**
adds the `configure` TUI; **Phase C** the bootstrap delivery.
