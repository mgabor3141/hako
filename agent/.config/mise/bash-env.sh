# Sourced via BASH_ENV in non-interactive bash (the agent's tool calls) so that
# mise-managed env (per-project [env], dotenv) is present outside interactive
# shells too. Interactive zsh uses `mise activate zsh` instead.
#
# We re-apply on cd because a non-interactive `bash -c 'cd X && cmd'` sources
# this at startup (before the cd), so without the wrappers cmd would see the
# starting dir's env, not X's. (This replaces what direnv used to do.)

# Put user-local bins on PATH (the inlined MCP CLI adapters symlink into
# ~/.local/bin at entrypoint; ADR-0013). Guarded against duplicates.
case ":$PATH:" in
  *":$HOME/.local/bin:"*) ;;
  *) export PATH="$HOME/.local/bin:$PATH" ;;
esac

command -v mise >/dev/null 2>&1 || return 0

_mise_hook() {
    local previous_exit_status=$?
    eval "$(mise hook-env -s bash 2>/dev/null)"
    return $previous_exit_status
}

# Load env for the initial working directory.
_mise_hook

# In non-interactive shells, no prompt fires; re-evaluate on directory change.
if [[ $- != *i* ]]; then
    cd()    { builtin cd    "$@" && _mise_hook; }
    pushd() { builtin pushd "$@" && _mise_hook; }
    popd()  { builtin popd  "$@" && _mise_hook; }
fi
