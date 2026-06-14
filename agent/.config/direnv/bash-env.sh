# Sourced via BASH_ENV in non-interactive bash (the agent's tool calls) so that
# direnv-managed variables are present outside interactive shells too.
#
# Interactive shells use `direnv hook zsh` instead (PROMPT_COMMAND/precmd).
# Non-interactive bash has no prompt, so we trigger the hook at startup and
# wrap the directory-changing builtins.
#
# The `BASH_ENV=` prefix prevents recursion: direnv spawns bash to evaluate
# .envrc, and that child must not re-enter this file.

command -v direnv >/dev/null 2>&1 || return 0

_direnv_hook() {
    local previous_exit_status=$?
    eval "$(BASH_ENV= direnv export bash 2>/dev/null)"
    return $previous_exit_status
}

# Load .envrc for the initial working directory.
_direnv_hook

# In non-interactive shells, PROMPT_COMMAND never fires; re-evaluate on cd.
if [[ $- != *i* ]]; then
    cd()    { builtin cd    "$@" && _direnv_hook; }
    pushd() { builtin pushd "$@" && _direnv_hook; }
    popd()  { builtin popd  "$@" && _direnv_hook; }
fi
