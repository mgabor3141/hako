#!/usr/bin/env bash
# tini is PID 1 and reaps children. This kicks off the home dev-toolchain
# reconcile (mise) as a *gmux session* -- so it shows up in the dashboard and
# can be watched -- then hands off to the CMD (gmuxd, the container's main
# service).
#
# The toolchain lives in the mounted home (not the image), so a cold clone must
# install it on first start. gmuxd has to be up before a session can launch, so
# we wait for it in the background, then run the reconcile from the home dir.
# Subsequent starts are fast (mise install is idempotent).
set -euo pipefail

state="$HOME/.local/state/hako"
mkdir -p "$state"
rm -f "$state/toolchain-ready" "$state/toolchain-failed"

# Put the inlined MCP CLI adapters on PATH. Sources live in the bind-mounted
# home (~/.agents/skills/<name>/<name>.ts); ~/.local/bin is on PATH. Re-linked
# every start so a cold clone has them. ADR-0013.
mkdir -p "$HOME/.local/bin"
for cli in "$HOME"/.agents/skills/*/; do
  name="$(basename "$cli")"
  [ -f "$cli$name.ts" ] && ln -sf "$cli$name.ts" "$HOME/.local/bin/$name"
done

if command -v mise >/dev/null 2>&1 && command -v gmux >/dev/null 2>&1; then
  (
    # wait for gmuxd (started below by exec) to be accepting sessions
    for _ in $(seq 1 150); do gmuxd status >/dev/null 2>&1 && break; sleep 0.2; done
    cd "$HOME"
    gmux bash -lc '
      if mise install --yes; then
        mise reshim >/dev/null 2>&1 || true
        touch "$HOME/.local/state/hako/toolchain-ready"
        echo; echo "hako: dev toolchain ready."
      else
        touch "$HOME/.local/state/hako/toolchain-failed"
        echo; echo "hako: toolchain install FAILED -- re-run: mise install"
      fi
    '
  ) &
fi

# Approval watcher (ADR-0010): when the gateway overlay mounts the approval
# scripts and points HAKO_APPROVAL_WATCH at the watcher, run it here -- it lives
# in this container because gmux's socket is local to it. It turns gated-call
# requests (dropped by the gateway hook in the shared /tmp/approvals volume)
# into interactive gmux y/N sessions. Absent the gateway, this is a no-op.
if [ -n "${HAKO_APPROVAL_WATCH:-}" ] && [ -x "${HAKO_APPROVAL_WATCH}" ] && command -v gmux >/dev/null 2>&1; then
  (
    for _ in $(seq 1 150); do gmuxd status >/dev/null 2>&1 && break; sleep 0.2; done
    exec "$HAKO_APPROVAL_WATCH"
  ) &
fi

exec "$@"
