# syntax=docker/dockerfile:1
#
# hako agent image. Everything hako needs lives OUTSIDE the user home (in /opt
# and /usr) so the entire home can be bind-mounted from the clone without
# shadowing anything. The mounted home is 100% the user's: pi config,
# projects, scratch. pi uses its default ~/.pi/agent (no env override).

FROM debian:stable-slim

ENV BUN_INSTALL=/opt/bun \
    PATH=/home/agent/.local/bin:/opt/bun/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl git ripgrep fd-find fzf less nano sudo tini unzip passwd \
      zsh zsh-autosuggestions zsh-syntax-highlighting \
 && rm -rf /var/lib/apt/lists/* \
 && ln -sf "$(command -v fdfind)" /usr/local/bin/fd \
 # zsh is the human's interactive shell; pi always runs /bin/bash itself.
 && useradd -m -s /usr/bin/zsh -u 1000 agent \
 && echo 'agent ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/agent

# Node.js — pi runs on node (its bin is `#!/usr/bin/env node`). NodeSource is
# portable across amd64/arm64. Lands in /usr (system), not the user home.
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
 && apt-get install -y --no-install-recommends nodejs \
 && rm -rf /var/lib/apt/lists/*

# bun + pi installed to /opt (outside the home), world-readable/executable.
RUN curl -fsSL https://bun.sh/install | bash \
 && /opt/bun/bin/bun install -g @earendil-works/pi-coding-agent \
 && chmod -R a+rX /opt/bun \
 # login shells re-run /etc/profile (resets PATH); keep /opt/bun/bin for them too
 && echo 'export PATH="$HOME/.local/bin:/opt/bun/bin:$PATH"' > /etc/profile.d/hako-path.sh

# gmux (connectivity): two static, dependency-free Go binaries. Installed to
# /usr/local/bin (OUTSIDE the home, or the bind mount would shadow them). The
# installer auto-detects amd64/arm64.
RUN curl -sSfL https://gmux.app/install.sh | GMUX_INSTALL_DIR=/usr/local/bin sh

# starship: a fast, git-aware prompt. Single static binary in /usr/local/bin
# (outside the home). Shell-agnostic; the shipped ~/.zshrc inits it.
RUN curl -sS https://starship.rs/install.sh | sh -s -- -y -b /usr/local/bin

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER agent
WORKDIR /home/agent
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/entrypoint.sh"]
# gmuxd runs in the foreground as the container's main service; `gmuxd run` is
# gmux's documented systemd/Docker entry point. Users exec in and `gmux pi`.
CMD ["gmuxd", "run"]
