# hako — an opinionated, bash-friendly zsh for newcomers.
# This ships in the mounted home: edit it freely, changes are live (no rebuild).

# --- bash compatibility, so commands the agent gives you paste-and-run ---
unsetopt nomatch             # a glob that matches nothing is passed through (like bash)
setopt   interactivecomments # allow `# comments` on the command line
unsetopt beep

# --- history: large, deduplicated, shared across sessions ---
HISTFILE=~/.zsh_history
HISTSIZE=50000
SAVEHIST=50000
setopt share_history hist_ignore_all_dups hist_ignore_space

# --- navigation: type a directory to cd into it; keep a stack ---
setopt auto_cd auto_pushd pushd_ignore_dups

# --- completion: case-insensitive, arrow-key menu ---
autoload -Uz compinit && compinit
zstyle ':completion:*' matcher-list 'm:{a-zA-Z}={A-Za-z}'
zstyle ':completion:*' menu select

# --- a few friendly, transparent aliases (plain commands underneath) ---
alias ll='ls -lah --color=auto'
alias la='ls -A --color=auto'
alias grep='grep --color=auto'

# --- fish-like ghost-text suggestions from your history ---
source /usr/share/zsh-autosuggestions/zsh-autosuggestions.zsh 2>/dev/null

# --- prompt: starship (git-aware, shows where you are) ---
eval "$(starship init zsh)"

# --- as-you-type syntax highlighting; must be sourced LAST ---
source /usr/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh 2>/dev/null
