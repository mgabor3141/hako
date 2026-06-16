#!/bin/sh
# Sealed boot for the gateway (ADR-0011). Used as the gateway entrypoint when the
# vault overlay is active. Starts with NO upstream token; blocks until
# hako-unlock writes the decrypted token to the tmpfs at /run/hako/token, then
# launches mcp-proxy with it in the environment. A gateway restart clears the
# tmpfs, so it comes back sealed -- re-unlock with `hako unlock`.
set -eu
tok="/run/hako/token"
echo "hako-gateway: SEALED -- waiting for unlock (run: hako unlock)" >&2
while [ ! -s "$tok" ]; do sleep 1; done
GITHUB_MCP_TOKEN="$(cat "$tok")"
export GITHUB_MCP_TOKEN
echo "hako-gateway: unsealed -- starting mcp-proxy" >&2
exec /usr/local/bin/mcp-proxy "$@"
