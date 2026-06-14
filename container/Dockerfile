# syntax=docker/dockerfile:1
#
# hako agent image. Everything hako needs lives OUTSIDE the user home (in /opt
# and /usr) so the entire home can be bind-mounted from the clone without
# shadowing anything. The mounted home is 100% the user's: pi config,
# projects, scratch. pi uses its default ~/.pi/agent (no env override).

FROM debian:stable-slim

ENV BUN_INSTALL=/opt/bun \
    PATH=/home/agent/.local/bin:/opt/bun/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    BASH_ENV=/home/agent/.config/direnv/bash-env.sh \
    DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl git ripgrep fd-find fzf less nano sudo tini unzip passwd \
      zsh zsh-autosuggestions zsh-syntax-highlighting fastfetch \
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

# Extra CLI goodies + general utilities + Python (agents lean on inline python).
# ffmpeg + imagemagick are the heavy part of this layer.
# --no-install-recommends stays (keeps pip from dragging in build-essential).
# Keep a base font so imagemagick can render text; patch for applying diffs.
# (Reading PDFs is covered by poppler-utils' pdftotext; imagemagick PDF
# rendering needs ghostscript ~ +50 pkgs, so add that in a fork if you want it.)
RUN apt-get update && apt-get install -y --no-install-recommends \
      eza bat jq shellcheck direnv zoxide patch \
      imagemagick fonts-dejavu-core poppler-utils sqlite3 \
      wget rsync zip zstd xz-utils file procps openssh-client \
      python3 python3-pip python3-venv \
 && rm -rf /var/lib/apt/lists/* \
 && ln -sf "$(command -v batcat)" /usr/local/bin/bat

# ffmpeg: a stripped static build (full codecs, no system deps) instead of the
# 207-package apt tree that drags in X11/GL we never use headless. ~1/3 the size
# of unstripped builds. ffmpeg + ffprobe only. Arch-detected.
RUN set -eux; \
    case "$(dpkg --print-architecture)" in \
      amd64) a=amd64 ;; arm64) a=arm64 ;; *) echo "unsupported arch"; exit 1 ;; \
    esac; \
    tmp="$(mktemp -d)"; \
    curl -fsSL "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-${a}-static.tar.xz" \
      | tar -xJ -C "$tmp"; \
    d="$(find "$tmp" -maxdepth 1 -type d -name 'ffmpeg-*-static')"; \
    install -m 0755 "$d/ffmpeg" "$d/ffprobe" /usr/local/bin/; \
    rm -rf "$tmp"; \
    ffmpeg -version | head -1

# jujutsu (jj): optional VCS, not in apt. Static musl release binary, arch-detected.
RUN set -eux; \
    case "$(dpkg --print-architecture)" in \
      amd64) a=x86_64 ;; arm64) a=aarch64 ;; *) echo "unsupported arch"; exit 1 ;; \
    esac; \
    v="$(curl -fsSL https://api.github.com/repos/jj-vcs/jj/releases/latest \
         | grep -oE '"tag_name": *"[^"]+"' | head -1 | grep -oE 'v[0-9][^"]*')"; \
    tmp="$(mktemp -d)"; \
    curl -fsSL "https://github.com/jj-vcs/jj/releases/download/${v}/jj-${v}-${a}-unknown-linux-musl.tar.gz" \
      | tar -xz -C "$tmp"; \
    install -m 0755 "$(find "$tmp" -maxdepth 2 -name jj -type f | head -1)" /usr/local/bin/jj; \
    rm -rf "$tmp"; \
    jj --version

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER agent
WORKDIR /home/agent
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/entrypoint.sh"]
# gmuxd runs in the foreground as the container's main service; `gmuxd run` is
# gmux's documented systemd/Docker entry point. Users exec in and `gmux pi`.
CMD ["gmuxd", "run"]
