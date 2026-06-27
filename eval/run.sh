#!/usr/bin/env bash
# Sweep the model x variant matrix over Terminal-Bench and write per-cell results.
#
# Prereqs (see README.md): Harbor installed in the active env, Docker running,
# provider API keys exported. Pins live here so a run is reproducible.
set -euo pipefail

cd "$(dirname "$0")"
export PYTHONPATH="$PWD/adapter:${PYTHONPATH:-}"

# --- pins (ADR-0008) --------------------------------------------------------
TB_DATASET="${TB_DATASET:-terminal-bench@2.0}"   # task set + version
PI_VERSION="${PI_VERSION:-latest}"               # pin to an exact version for real runs
HAKO_REF="${HAKO_REF:-feat/eval-harness}"        # config under test
N_ATTEMPTS="${N_ATTEMPTS:-3}"                     # trials per task (stochastic)

# --- matrix -----------------------------------------------------------------
MODELS=(${MODELS:-anthropic/claude-sonnet-4-5})
VARIANTS=(${VARIANTS:-baseline goal})
# Optional: TASKS="hello-world foo" restricts the run (smoke-test before sweeping).
TASK_ARGS=()
for t in ${TASKS:-}; do TASK_ARGS+=(--task-id "$t"); done

STAMP="$(date +%Y%m%d-%H%M%S)"
for model in "${MODELS[@]}"; do
  for variant in "${VARIANTS[@]}"; do
    goal=false; [ "$variant" = "goal" ] && goal=true
    safe="${model//\//_}.${variant}"
    out="results/${STAMP}/${safe}"
    echo ">>> $model / $variant -> $out"
    tb run \
      --dataset "$TB_DATASET" \
      --import-path hako_agent:HakoAgent \
      --agent-kwarg "model_name=$model" \
      --agent-kwarg "goal=$goal" \
      --agent-kwarg "pi_version=$PI_VERSION" \
      --agent-kwarg "hako_ref=$HAKO_REF" \
      --n-attempts "$N_ATTEMPTS" \
      --output-path "$out" \
      "${TASK_ARGS[@]}"
  done
done

echo "Done. Results under results/${STAMP}/ — summarise into BASELINE.md if this is a reference run."
