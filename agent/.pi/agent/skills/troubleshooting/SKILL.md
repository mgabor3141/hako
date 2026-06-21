---
name: troubleshooting
description: >
  Diagnose failures of the hako environment itself -- a tool or skill erroring,
  auth/"unconfigured" errors, approval timeouts, sidecar or gateway connection
  failures, or git push being rejected. Says where to look and splits what you
  can fix in the sandbox from what needs the human (config, credentials, host).
  Not for ordinary bugs in your own code.
---

# troubleshooting

You run in a sandboxed container with **no credentials**. The human controls the
host `hako` launcher, the `hako.toml` config, and the locked vault. That split
decides who fixes what -- diagnose first, then act or escalate.

## You can fix these (inside the container)

- **A command/tool is missing.** The toolchain installs at startup into the home.
  Check `~/.local/state/hako/toolchain-ready` (vs `toolchain-failed`); if it
  failed or is absent, run `mise install` then `mise reshim`.
- **Python `ImportError`.** `python -m pip install <pkg>` (the office libs are
  pinned in `~/.default-python-packages`). It persists in the home.
- **A skill's CLI isn't on PATH but `~/.agents/skills/<name>/` exists.** CLIs are
  linked at startup from `<name>.ts` into `~/.local/bin/`; re-link it, or run the
  `.ts` directly.

## Ask the human (you cannot reach these)

- **Auth / "unconfigured" error from a skill** (e.g. `github` exits **3**). The
  credential lives in the vault, added on the host with `hako auth <name>`. You
  hold none and can't add it -- say which integration needs auth.
- **Approval timed out** (e.g. `github` exits **4**). A gated write is waiting for
  the human to approve in gmux. Ask them to approve.
- **A skill is missing entirely** (no `~/.agents/skills/<name>/`). It's disabled
  in `hako.toml`. Ask them to enable it (`hako configure`, then `hako up`); you
  can't assemble the stack.
- **Sidecar connection refused** (e.g. websearch `http://websearch:8080`, webview
  `http://webview:11235`). The sidecar isn't running -- host-side. Ask them to
  enable it and `hako up`.
- **`git push` rejected (auth).** Needs a human-set-up push credential: suggest
  `/setup-git` or `docs/git.md`. Never fetch or invent a token yourself.
- **Anything host-side** -- image rebuild (`hako up --build`), gmux connectivity,
  the vault passphrase, host networking. The launcher runs on the host; describe
  the symptom and ask.

## Where to look

- Toolchain state: `~/.local/state/hako/toolchain-{ready,failed}`.
- Enabled skills: `ls ~/.agents/skills/`; integration manifests at
  `/opt/hako/integrations/`.
- A gateway skill's endpoint: env `<NAME>_MCP_URL`.
- `github` exit codes: 2 unsupported, 3 unconfigured, 4 approval timeout.
