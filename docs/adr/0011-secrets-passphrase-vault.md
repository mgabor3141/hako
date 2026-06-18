# ADR-0011: Secrets — passphrase-vault by default

- **Status:** Accepted — 2026-06-16. **Built, in-process:** one `vault/secrets.age`
  under a single global passphrase; `hako unlock` decrypts it on the host
  (`filippo.io/age` + locked memory) and pipes the env into the gateway's tmpfs
  (the gateway no longer decrypts). Validated against real GitHub.

## Context
The gateway holds the real upstream credentials (ADR-0002/0007). Since it's a
container, it can't read host credential helpers — the secret must reach it
somehow. The target audience is someone wiring an agent into their Gmail for the
first time, so the **default must be as safe as possible**, with lighter options
for people who know what they're doing.

## Decision
The **passphrase-vault** (the seal/unseal model, à la Vault) is the mechanism.
Secrets live **`age`-encrypted in a mounted folder** as a **single vault under
one global passphrase** (one unlock for everything — per-secret passphrases just
push people to one weak shared password); the gateway boots **sealed** and blocks
until unsealed. `hako unlock` (host) prompts for the passphrase (masked) and
**decrypts the vault in-process** (`filippo.io/age` + locked, zeroized memory —
no subprocess, no pty), then pipes the resulting env into the gateway's tmpfs at
**`/run/hako/env`** (so the secret never appears in `docker inspect`); the
gateway sources it and launches mcp-proxy. The gateway never sees the passphrase
or the ciphertext. Restart clears the tmpfs and re-seals. The plaintext is never
on the host disk, the encrypted blob is useless without the passphrase, and the
passphrase is never stored.

The **source of a secret is never prescribed**: the vault is what hako ships and
wires, but you can populate it from a secret-manager CLI (`op run`, `vault`,
keychain). Encrypted-at-rest alternatives (a Docker Secrets Engine / OS keychain,
`se://`) are a reasonable opt-down for those who know what they're doing; hako
does not build them in. (An earlier draft framed these as built-in "tiers"; in
practice only the vault was implemented.)

## Consequences
Strongest-by-default protects against other users, **same-user non-root
processes** (process memory in the gateway's namespace; default `ptrace_scope`),
and disk-at-rest (backups, theft, stray copy). It does **not** beat root or the
Docker daemon — no software-only scheme does without a TPM/enclave. Cost: a
**passphrase on every gateway restart** (the deliberate seal/unseal tradeoff);
an `se://`-style keychain would trade that for login-session convenience. Built
in-process in the Go launcher (`filippo.io/age`); no separate decrypt utility or
pty hack, and no `age` in the gateway image.
