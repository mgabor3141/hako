# ADR-0010: Tool-call approval is a swappable hook

- **Status:** Accepted — 2026-06-16 (amended 2026-06-16: the default channel is
  no longer notify-send — see Decision). `callHook` is wired and interactively
  verified in the tracer — approve / deny / policy-deny, via a real human click.

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

**The shipped hook stays OS-agnostic.** The channel is the operator's choice, but
hako's default must not assume an OS: host/desktop-specific wiring lives in a
host-side responder (the launcher's job, ADR-0012) and in a local, gitignored
override (`compose.override.yaml` + `gateway/hooks.local/`) — never in the
shipped image, compose, or hook.

**Intended default: a persistent approval surface, not a fire-and-forget toast.**
Dismiss an OS notification by accident and the request can only time out. The
target is a **persistent UI — a gmux TUI reading persisted approval state** the
human can reopen and act on at leisure (gmux needs a notify primitive for this).
Until it exists, the shipped "ask" branch is a safe placeholder: deny unless an
explicit allow toggle is set.

## Consequences
The hook is one swappable command, so any channel drops in: a local
`notify-send --wait --action` desktop toast (**proven end-to-end** on a Linux
host via the gitignored override — gateway run as the host UID for D-Bus auth,
host session bus mounted — but **not shipped**, since toasts lack persistence),
**`ntfy` push** to a phone (good for WSL2/headless), a webhook, or the persistent
TUI. We ship examples, not a baked-in channel.

Because notify-send is no longer the default, the **gateway image ships no
libnotify** and stays minimal; desktop toasts are a documented, local-only
option, so the Linux-only D-Bus-mount boundary is a property of *that option*,
not of hako. A future ADR may split out the persistent-surface design once gmux
grows the primitive.
