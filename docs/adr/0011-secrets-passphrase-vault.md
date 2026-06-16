# ADR-0011: Secrets — passphrase-vault by default

- **Status:** Accepted — 2026-06-16. Vault tracer **built**: `age` blob -> sealed
  gateway -> `hako unlock` (passphrase over stdin) -> token in tmpfs, validated
  against real GitHub. `.env` (tier 3) remains the no-vault fallback.

## Context
The gateway holds the real upstream credentials (ADR-0002/0007). Since it's a
container, it can't read host credential helpers — the secret must reach it
somehow. The target audience is someone wiring an agent into their Gmail for the
first time, so the **default must be as safe as possible**, with lighter options
for people who know what they're doing.

## Decision
Three tiers, all converging on the gateway reading the secret from a tmpfs file
at **`/run/secrets/<name>`** (so it never appears in `docker inspect`):

1. **Default — passphrase-vault (the seal/unseal model, à la Vault).** Secrets
   live **`age`-encrypted in a mounted folder**; the gateway boots **sealed** and
   blocks until unsealed. `hako up` prompts for the passphrase (masked input)
   and pipes it to `docker exec -i gateway hako-unlock` over **stdin** — never in
   args/env/history. The gateway decrypts into **memory** and only then launches
   mcp-proxy. Restart re-seals. (age refuses a non-tty passphrase, so `hako-unlock`
   pty-wraps it via `script`.) The plaintext is never on the host disk, the
   encrypted blob is useless without the passphrase, and the passphrase is never
   stored. The gateway *is* the in-memory secret agent (ssh-agent/gpg-agent
   pattern): everything goes through it, nothing else sees the credential.
2. **`se://` (Docker Secrets Engine / OS keychain)** — documented opt-down.
   Lower friction (rides the login session), encrypted at rest, no file/env/
   history. Bundled with Docker Desktop; CE injection is on Docker's roadmap.
3. **`0600` gitignored file** — simplest fallback, plaintext at rest.

The **source is never prescribed** — a secret-manager CLI (`op run`, `vault`,
keychain) or a file can populate any tier; we document, we don't mandate.

## Consequences
Strongest-by-default protects against other users, **same-user non-root
processes** (process memory in the gateway's namespace; default `ptrace_scope`),
and disk-at-rest (backups, theft, stray copy). It does **not** beat root or the
Docker daemon — no software-only scheme does without a TPM/enclave. Cost: a
**passphrase on every gateway restart** (the deliberate seal/unseal tradeoff);
`se://` trades that for login-session convenience. Implement the vault with
`age` + a `hako-unlock` utility after the tracer proves the loop.
