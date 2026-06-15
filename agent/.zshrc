# hako — an opinionated, bash-friendly zsh for newcomers.
# This ships in the mounted home: edit it freely, changes are live (no rebuild).

# --- bash compatibility, so commands the agent gives you paste-and-run ---
unsetopt nomatch             # a glob that matches nothing is passed through (like bash)
setopt   interactivecomments # allow `# comments` on the command line
unsetopt beep

# --- welcome banner on a fresh terminal (skip nested/subshells) ---
if [[ $SHLVL -le 1 ]]; then
  command -v fastfetch >/dev/null && fastfetch
  # first-run hint while the dev toolchain installs (as a gmux session)
  if [[ -f ~/.local/state/hako/toolchain-failed ]]; then
    print -P "%F{red}hako:%f toolchain install failed — re-run %F{cyan}mise install%f"
  elif [[ ! -f ~/.local/state/hako/toolchain-ready ]]; then
    print -P "%F{yellow}hako:%f installing the dev toolchain (first run); watch it in the gmux dashboard. Tools appear as they finish."
  fi
  print -P "%F{8}tip: type %F{cyan}hako%F{8} for keybindings & handy commands%f"
fi

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

# --- eza: a friendlier ls (git-aware, tree, icons). Matches the host dotfiles.
# Icons need a Nerd Font in your terminal (gmux Nerd Font support is planned). ---
alias ls='eza -al --icons --group-directories-first --git'      # long + all (preferred)
alias ll='eza -l  --icons --group-directories-first --git'      # long, no dotfiles
alias la='eza -a  --icons --group-directories-first'            # all, grid
alias lt='eza -aT --icons --level=2 --group-directories-first'  # tree, 2 levels
alias l.="eza -a | grep -e '^\.'"                                # dotfiles only
alias grep='grep --color=auto'

# launch the agent through gmux, so the session shows up in the dashboard
alias pi='gmux pi'

# --- hako: a quick reference for the goodies that aren't easy to discover ---
hako() {
  print -P '%F{cyan}%Bhako quick reference%b%f   (the non-obvious goodies)

%F{yellow}search & history (fzf)%f
  %BCtrl-R%b    fuzzy-search your command history
  %BCtrl-T%b    fuzzy-pick a file path onto the command line
  %BAlt-C%b     fuzzy-cd into a subdirectory
  %B->%b / End  accept the grey autosuggestion (from history)

%F{yellow}getting around%f
  %Bz <name>%b    jump to a directory you visit often (zoxide)
  %Bzi%b          pick the directory interactively
  %B<dir>%b       just type a directory name to cd into it (auto_cd)
  %Bcd -<Tab>%b   pick from the directory stack

%F{yellow}listing (eza) & reading%f
  %Bls ll la%b    long+all / long / all      %Blt%b tree   %Bl.%b dotfiles
  %Bbat FILE%b    syntax-highlighted cat     %Bman CMD%b colorized man pages

%F{yellow}agent & tooling%f
  %Bpi%b          launch the coding agent (opens a gmux session)
  %Bmise ls%b     list installed tools   %Bmise use -g <tool>@<ver>%b  add/bump one
  %Bjj st%b       version control (jujutsu)

%F{8}browser dashboard: http://localhost:8791   (token: gmuxd auth)%f'
}

# --- bat: syntax-highlighted, colored man pages ---
export MANROFFOPT="-c"
export MANPAGER="sh -c 'col -bx | bat -l man -p'"

# --- mise: dev toolchain manager (puts node/bun/python/pi/CLI tools on PATH) ---
command -v mise >/dev/null && eval "$(mise activate zsh)"

# --- tool integrations (no-op if a tool is missing; mise handles project env) ---
command -v zoxide >/dev/null && eval "$(zoxide init zsh)"            # `z <dir>` jumps
command -v fzf    >/dev/null && source <(fzf --zsh) 2>/dev/null      # Ctrl-R history, Ctrl-T files, Alt-C cd

# --- fish-like ghost-text suggestions from your history ---
source /usr/share/zsh-autosuggestions/zsh-autosuggestions.zsh 2>/dev/null

# --- prompt: starship (git-aware, shows where you are) ---
eval "$(starship init zsh)"

# --- as-you-type syntax highlighting; must be sourced LAST ---
source /usr/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh 2>/dev/null
