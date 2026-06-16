#!/bin/sh
# Runs INSIDE a gmux session in the hako container (spawned by watch.sh). Shows
# the gated call and asks the human y/N, writing the verdict the gateway hook is
# blocking on. ADR-0010.
set -eu
id="${1:?usage: prompt.sh <id>}"
dir="${HAKO_APPROVALS_DIR:-/tmp/approvals}"
req="$dir/$id.json"
[ -f "$req" ] || { echo "hako: no such approval request: $id"; exit 1; }

clear 2>/dev/null || true
echo "============================  hako approval  ============================"
if command -v jq >/dev/null 2>&1; then
  printf 'server : %s\n' "$(jq -r '.server' "$req")"
  printf 'tool   : %s\n' "$(jq -r '.params.name' "$req")"
  echo   'args   :'
  jq -C '.params.arguments' "$req" 2>/dev/null | sed 's/^/  /'
else
  cat "$req"
fi
echo "========================================================================="
printf 'Approve this call? [y/N] '
read -r ans
case "$ans" in
  y|Y|yes|YES) printf approve > "$dir/$id.verdict"; echo; echo "APPROVED." ;;
  *)           printf deny    > "$dir/$id.verdict"; echo; echo "DENIED." ;;
esac
