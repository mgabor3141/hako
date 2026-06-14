# syntax=docker/dockerfile:1
#
# hako agent image. Everything hako needs lives OUTSIDE the user home (/opt,
# /usr, /nix-free) so the entire home can be bind-mounted from the clone
# without shadowing anything. The mounted home is 100% the user's: pi config,
# projects, scratch. pi uses its default ~/.pi/agent (no env override).

FROM debian:stable-slim

ENV BUN_INSTALL=/opt/bun \
    PATH=/opt/bun/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl git ripgrep fd-find fzf less nano sudo tini unzip passwd \
 && rm -rf /var/lib/apt/lists/* \
 && ln -sf "$(command -v fdfind)" /usr/local/bin/fd \
 && useradd -m -s /bin/bash -u 1000 agent \
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
 && echo 'export PATH=/opt/bun/bin:$PATH' > /etc/profile.d/hako-path.sh

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER agent
WORKDIR /home/agent
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/entrypoint.sh"]
CMD ["sleep", "infinity"]
