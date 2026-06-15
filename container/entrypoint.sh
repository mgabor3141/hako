#!/usr/bin/env bash
# tini is PID 1 and reaps children. This wrapper reconciles the home dev
# toolchain (mise) to the committed lockfile, then hands off to the CMD, which
# runs gmuxd in the foreground as the container's main service.
#
# The toolchain lives in the mounted home (not the image), so a cold clone must
# install it on first start. We do that in the BACKGROUND so the gmux web UI
# comes up immediately; a readiness flag lets the shell banner report progress.
# Subsequent starts are fast (mise install is idempotent).
set -euo pipefail

state="$HOME/.local/state/hako"
mkdir -p "$state"
rm -f "$state/toolchain-ready" "$state/toolchain-failed"

if command -v mise >/dev/null 2>&1; then
  (
    if mise install --yes >"$state/bootstrap.log" 2>&1; then
      mise reshim >/dev/null 2>&1 || true
      touch "$state/toolchain-ready"
    else
      touch "$state/toolchain-failed"
    fi
  ) &
fi

exec "$@"
