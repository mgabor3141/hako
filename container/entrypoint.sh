#!/usr/bin/env bash
# tini is PID 1. Start gmuxd in the background if it's installed (connectivity
# layer, added later), then keep the container alive. Nothing here touches the
# bind-mounted home beyond what the agent does itself.
set -euo pipefail

if command -v gmuxd >/dev/null 2>&1; then
  gmuxd start || true
fi

exec "$@"
