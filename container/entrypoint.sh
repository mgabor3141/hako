#!/usr/bin/env bash
# tini is PID 1 and reaps the agent's child processes. This wrapper is a seam
# for seeding mutable home state later; for now it just hands off to the CMD,
# which runs gmuxd in the foreground as the container's main service.
set -euo pipefail

exec "$@"
