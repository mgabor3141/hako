# Driving pi from an editor (ACP)

hako can be driven from an external editor that speaks the
[Agent Client Protocol](https://agentclientprotocol.com) (ACP) -- e.g. Zed. You
get the editor's chat/diff UI, but the agent stays **inside the container**: same
sandbox, same credential isolation, same gmux approval gate.

## How it works

```
Zed (host)  <--ACP JSON-RPC / stdio-->  docker exec -i hako pi-acp  -->  pi --mode rpc
                                              (both run in the container)
```

[`pi-acp`](https://github.com/svkozak/pi-acp) is a small ACP adapter (shipped as
a pinned mise tool, see `agent/.config/mise/config.toml`). It talks ACP over
stdio to the editor and spawns `pi --mode rpc` next to it. `docker exec -i` (no
TTY) is just a clean stdio pipe across the container boundary.

Crucially, **pi-acp does no ACP filesystem/terminal delegation** -- pi reads,
writes, and runs commands *locally, in the container*. The editor only receives
tool-call events, structured diffs, and file paths. So nothing about the sandbox
changes: pi still holds zero credentials, MCP tool calls still route through the
gateway + approval gate, and the host only needs `docker exec` (which it already
has to run hako at all).

## Setup

1. Bring the stack up (`./hako up`). `pi-acp` is on the container's PATH; no login
   shell is involved, so stdout stays a clean JSON-RPC stream.

2. Point your editor at the container. In Zed's `settings.json`:

   ```json
   "agent_servers": {
     "hako": {
       "type": "custom",
       "command": "docker",
       "args": ["exec", "-i", "hako", "pi-acp"],
       "env": {}
     }
   }
   ```

   (`hako` is the container name. Add `"PI_ACP_ENABLE_EMBEDDED_CONTEXT": "true"`
   to `env` if you want the editor to send embedded resource blocks.)

3. Open the agent panel in Zed and select **hako**. pi's own auth is unchanged --
   if pi needs an API key, Zed's "Authenticate" banner runs `pi --terminal-login`
   in the container.

## The one rule: matching paths

The editor sends the **host** path of the project you have open as the session's
working directory, and pi-acp `cd`s pi there. That path must exist *at the same
absolute path inside the container*, or pi has nowhere to work.

So: **bind-mount the project you want the sandboxed agent to edit at an identical
source:target path.** Drop a `compose.override.yaml` next to `compose.yaml`
(gitignored -- it's host-specific):

```yaml
services:
  hako:
    volumes:
      # identical host path : container path -- NOT /home/agent/...
      - /home/you/dev/myproject:/home/you/dev/myproject
```

Then open `/home/you/dev/myproject` in Zed. Now everything lines up:

- pi's cwd is valid;
- edits land on the bind-mounted host files (you see them on the host immediately);
- Zed's inline diffs **and** click-to-open follow-along resolve, because the paths
  pi emits are real on both sides.

You're widening the sandbox to that one project -- which is the point -- while the
rest of your machine and your credentials stay out of reach. (The agent runs as a
uid that matches your host user, same as the `./agent` home mount, so it can write
the mounted files.)

## Notes / limitations

- `pi-acp` is a pre-1.0, MVP-stage adapter; expect occasional breaking changes.
  Bump it deliberately like any mise tool: `mise use -g npm:pi-acp@<version>`
  inside the container, then commit the `mise.lock` diff.
- MCP servers offered *over ACP* by the editor are not wired into pi. That's fine
  here -- pi's MCP integrations (github, websearch, ...) come from hako's own
  gateway and work normally regardless of who's driving pi.
- Don't run the adapter through a login shell (`bash -lc`): the startup banner
  would corrupt the JSON-RPC stream. The plain `docker exec -i hako pi-acp` above
  avoids that by relying on the mise shims already being on the default PATH.
