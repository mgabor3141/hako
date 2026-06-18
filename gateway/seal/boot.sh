#!/bin/sh
# Sealed boot for the gateway (ADR-0011). Used as the gateway entrypoint when the
# vault overlay is active. Starts with NO secrets; blocks until `hako unlock`
# (running on the host) decrypts the vault and pipes the env into the tmpfs at
# /run/hako/env. It then sources that env and launches mcp-proxy. Decryption
# happens on the host (filippo.io/age, in locked memory) -- this container never
# sees the passphrase or the ciphertext. A restart clears the tmpfs -> re-sealed.
set -eu
env="/run/hako/env"
echo "hako-gateway: SEALED -- waiting for unlock (run: hako unlock)" >&2
while [ ! -s "$env" ]; do sleep 1; done
# a runtime file, not present at lint time:
# shellcheck source=/dev/null
. "$env"   # export KEY='value' lines written by the host
echo "hako-gateway: unsealed -- starting mcp-proxy" >&2
exec /usr/local/bin/mcp-proxy "$@"
