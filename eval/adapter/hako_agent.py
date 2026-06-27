"""Harbor installed-agent adapter that benchmarks hako's *pi config* on Terminal-Bench.

What is under test is hako's shipped pi configuration — the agent home at
`agent/.pi/agent/` (AGENTS.md, settings.json, prompts/, skills/, and the
extensions, including goal.ts) — driven by `pi -p`. The container the task runs
in is Terminal-Bench's own per-task image; we only install pi + the config home
into it. The hako *container* is deliberately out of scope (each task brings its
own environment); the behaviourally load-bearing part is the config home.

Two cells, selected by the `goal` kwarg:
  - baseline:  pi -p "<instruction>"
  - +goal:     same, but PI_GOAL_AUTOSTART=1 so goal.ts runs its diligent-user
               loop headlessly (no `/goal` to type in print mode).

Run (from eval/adapter, with Harbor installed in the same env):
  tb run --import-path hako_agent:HakoAgent \
         --agent-kwarg model_name=anthropic/claude-sonnet-4-5 \
         --agent-kwarg goal=true \
         --task-id hello-world
"""

import os
import shlex
from pathlib import Path

from terminal_bench.agents.installed_agents.abstract_installed_agent import (
    AbstractInstalledAgent,
)
from terminal_bench.terminal.models import TerminalCommand

# provider prefix (in "provider/model") -> the env var pi reads the key from.
_PROVIDER_KEY_ENV = {
    "anthropic": "ANTHROPIC_API_KEY",
    "openai": "OPENAI_API_KEY",
    "google": "GEMINI_API_KEY",
    "gemini": "GEMINI_API_KEY",
    "openrouter": "OPENROUTER_API_KEY",
    "xai": "XAI_API_KEY",
    "groq": "GROQ_API_KEY",
}


def _as_bool(v) -> bool:
    if isinstance(v, bool):
        return v
    return str(v).strip().lower() in {"1", "true", "yes", "on"}


class HakoAgent(AbstractInstalledAgent):
    """hako's pi config, installed into the task container and run with `pi -p`."""

    def __init__(
        self,
        model_name: str | None = None,
        goal: bool | str = False,
        goal_text: str = "",
        proxy_model: str | None = None,
        hako_ref: str = "feat/eval-harness",
        pi_version: str = "latest",
        *args,
        **kwargs,
    ):
        super().__init__(*args, **kwargs)
        if not model_name:
            raise ValueError("model_name is required, e.g. anthropic/claude-sonnet-4-5")
        self._model_name = model_name
        self._goal = _as_bool(goal)
        self._goal_text = goal_text
        self._proxy_model = proxy_model
        self._hako_ref = hako_ref
        self._pi_version = pi_version

    @staticmethod
    def name() -> str:
        return "hako"

    @property
    def _provider(self) -> str:
        return self._model_name.split("/", 1)[0].lower()

    @property
    def _env(self) -> dict[str, str]:
        key_env = _PROVIDER_KEY_ENV.get(self._provider)
        if not key_env:
            raise ValueError(
                f"unknown provider '{self._provider}' in model '{self._model_name}'; "
                f"add it to _PROVIDER_KEY_ENV"
            )
        if key_env not in os.environ:
            raise ValueError(f"missing {key_env} in environment for {self._model_name}")
        env = {key_env: os.environ[key_env]}
        if self._goal:
            env["PI_GOAL_AUTOSTART"] = "1"
            env["PI_GOAL"] = self._goal_text
            if self._proxy_model:
                # goal.ts picks its proxy from the first PI_LIBRARIAN_MODELS token.
                env["PI_LIBRARIAN_MODELS"] = self._proxy_model
        return env

    @property
    def _install_agent_script_path(self) -> Path:
        return self._get_templated_script_path("hako-setup.sh.j2")

    def _get_template_variables(self) -> dict[str, str]:
        return {"pi_version": self._pi_version, "hako_ref": self._hako_ref}

    def _run_agent_commands(self, instruction: str) -> list[TerminalCommand]:
        cmd = (
            f"pi -p {shlex.quote(instruction)} "
            f"--model {shlex.quote(self._model_name)} --no-session"
        )
        return [
            TerminalCommand(
                command=cmd,
                min_timeout_sec=0.0,
                max_timeout_sec=float("inf"),  # let the task-level timeout govern
                block=True,
                append_enter=True,
            ),
        ]
