# syntax=docker/dockerfile:1
#
# hako agent image -- the OS (authoritative) tier.
#
# This image is deliberately thin: base system, shells, sudo, tini, the gmux
# connectivity layer, ffmpeg, and the `mise` tool manager. The DEV TOOLCHAIN
# (node, bun, python, pi, and the CLI tools) is NOT baked here -- mise installs
# it into the mounted home at container start, pinned by agent/.config/mise/
# config.lock. See container/entrypoint.sh and docs/adr/0004.
#
# Everything baked here lives OUTSIDE the home, so the whole-home bind mount
# shadows nothing. Every external download is version- and checksum-pinned.

FROM debian:trixie-slim

# Pinned, checksum-verified downloads. Bump deliberately (a visible commit).
ARG MISE_VERSION=v2026.6.10
ARG MISE_SHA256_X64=656131b5d995493c18b06c2135afda193f94a1848e0a4d1cb07a1161859d0c80
ARG MISE_SHA256_ARM64=f39ef86c888efe7323c1e510a8d14614af43d030acb3d56cb1af47e66b87ef28

ARG GMUX_VERSION=v1.6.0
ARG GMUX_SHA256_AMD64=fa6732e3703dbab789ab3dc452d43bcdb75871cad4ae5b1f6c4b86aa651af378
ARG GMUX_SHA256_ARM64=5cd6182e7c6b991e989ca3ebcb3fd144fbac978ec28ba017597581d4bf76fd2c

ARG FFMPEG_SHA256_AMD64=abda8d77ce8309141f83ab8edf0596834087c52467f6badf376a6a2a4c87cf67
ARG FFMPEG_SHA256_ARM64=f4149bb2b0784e30e99bdda85471c9b5930d3402014e934a5098b41d0f7201b1

# Shims dir first: the agent's non-interactive bash finds mise-managed tools
# without activation. Interactive zsh additionally runs `mise activate`.
ENV PATH=/home/agent/.local/share/mise/shims:/home/agent/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    BASH_ENV=/home/agent/.config/mise/bash-env.sh \
    DEBIAN_FRONTEND=noninteractive

# OS tier: base + shells + a few expected system utilities. No dev toolchain.
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl git xz-utils less sudo tini passwd \
      procps file patch \
      zsh zsh-autosuggestions zsh-syntax-highlighting fastfetch \
 && rm -rf /var/lib/apt/lists/* \
 # zsh is the human's interactive shell; pi always runs /bin/bash itself.
 && useradd -m -s /usr/bin/zsh -u 1000 agent \
 && echo 'agent ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/agent

# mise: single pinned, checksum-verified binary. Manages the home toolchain.
RUN set -eux; \
    case "$(dpkg --print-architecture)" in \
      amd64) a=x64;   sha="$MISE_SHA256_X64" ;; \
      arm64) a=arm64; sha="$MISE_SHA256_ARM64" ;; \
      *) echo "unsupported arch"; exit 1 ;; \
    esac; \
    curl -fsSL -o /usr/local/bin/mise "https://github.com/jdx/mise/releases/download/${MISE_VERSION}/mise-${MISE_VERSION}-linux-${a}"; \
    echo "${sha}  /usr/local/bin/mise" | sha256sum -c -; \
    chmod 0755 /usr/local/bin/mise; \
    mise --version

# gmux (connectivity): pinned, checksum-verified release archive (gmux + gmuxd).
# gmuxd is the container's main service; it can only be updated by rebuild.
RUN set -eux; \
    arch="$(dpkg --print-architecture)"; \
    case "$arch" in \
      amd64) sha="$GMUX_SHA256_AMD64" ;; \
      arm64) sha="$GMUX_SHA256_ARM64" ;; \
      *) echo "unsupported arch"; exit 1 ;; \
    esac; \
    tmp="$(mktemp -d)"; \
    curl -fsSL -o "$tmp/gmux.tgz" "https://github.com/gmuxapp/gmux/releases/download/${GMUX_VERSION}/gmux_${GMUX_VERSION#v}_linux_${arch}.tar.gz"; \
    echo "${sha}  $tmp/gmux.tgz" | sha256sum -c -; \
    tar -xzf "$tmp/gmux.tgz" -C "$tmp"; \
    install -m 0755 "$tmp/gmux" "$tmp/gmuxd" /usr/local/bin/; \
    rm -rf "$tmp"; \
    gmuxd version

# ffmpeg: stripped static build (OS tier -- it is heavy and awkward in mise).
# ffmpeg + ffprobe only. Pinned by hash; bump when johnvansickle rolls a release.
RUN set -eux; \
    case "$(dpkg --print-architecture)" in \
      amd64) a=amd64; sha="$FFMPEG_SHA256_AMD64" ;; \
      arm64) a=arm64; sha="$FFMPEG_SHA256_ARM64" ;; \
      *) echo "unsupported arch"; exit 1 ;; \
    esac; \
    tmp="$(mktemp -d)"; \
    curl -fsSL -o "$tmp/ff.tar.xz" "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-${a}-static.tar.xz"; \
    echo "${sha}  $tmp/ff.tar.xz" | sha256sum -c -; \
    tar -xJf "$tmp/ff.tar.xz" -C "$tmp"; \
    d="$(find "$tmp" -maxdepth 1 -type d -name 'ffmpeg-*-static')"; \
    install -m 0755 "$d/ffmpeg" "$d/ffprobe" /usr/local/bin/; \
    rm -rf "$tmp"; \
    ffmpeg -version | head -1

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER agent
WORKDIR /home/agent
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/entrypoint.sh"]
# gmuxd runs in the foreground as the container's main service. The entrypoint
# kicks off the home-toolchain install in the background first.
CMD ["gmuxd", "run"]
