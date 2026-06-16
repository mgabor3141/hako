# ADR-0010: Tool-call approval is a swappable hook

- **Status:** Accepted — 2026-06-16. `callHook` is built (in the gateway fork);
  not yet wired into hako.

## Context
A sandboxed agent reaching real systems through the gateway (ADR-0007) needs a
human in the loop for sensitive actions (sending mail, merging PRs). Where and
how does approval happen, given the gateway is a Linux container?

## Decision
The gateway's per-server **`callHook`** gates only the tools listed in
`requireFor` (reads pass straight through). For a gated call it runs a command
with the request JSON on stdin (+ `MCP_SERVER`/`MCP_TOOL`): **exit 0 approves;
any non-zero, spawn error, or timeout denies (fail-closed)** with a readable MCP
error the model can react to. The command is **arbitrary, set in the gateway's
config**, so the approval channel is the operator's choice.

**Default channel: a Linux desktop notification with Approve/Deny**
(`notify-send --wait --action`), wired by mounting the host session bus
(`$XDG_RUNTIME_DIR/bus`) into the **gateway** container only (the trusted,
cred-holding component — never the agent). Click-to-approve, no host process, no
extra port.

## Consequences
Approval is one swappable command, so users not on a Linux desktop drop in their
own: **`ntfy` push** (works from any container, notifies a phone, action button
posts back — good for WSL2), **host-process mode** for macOS Notification Center
(ADR-0007), or a webhook/TUI. We ship examples, not a baked-in channel.

Container boundary, stated plainly: a Linux gateway container can raise a host
notification only when the host is **Linux** (D-Bus mount); macOS/Windows hosts
need a host helper or host-process mode. The gateway image gains `notify-send` +
libnotify (a few MB) for the default — still tiny next to a node+python image.
`--wait`/`--action` support depends on the notification daemon (GNOME/KDE fine).
