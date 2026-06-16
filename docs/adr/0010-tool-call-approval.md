# ADR-0010: Tool-call approval is a swappable hook

- **Status:** Accepted — 2026-06-16 (amended 2026-06-16: the default channel is
  the shipped default is a gmux approval session — see Decision). `callHook` is
  wired and interactively verified in the tracer (approve / deny / policy-deny,
  via a real human answering in the browser).

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

**Shipped default: an interactive gmux session.** A fire-and-forget OS toast is
inadequate — dismiss one by accident and the request can only time out. Instead
the hook drops the request into a shared `/tmp/approvals` volume and blocks on a
verdict file; a **hako-side watcher** (gmux's socket is local to the agent
container, not the gateway) turns each request into a `y/N` gmux session that
**persists in the dashboard until answered**, OS-agnostic and browser-first.
Exited prompts linger with their verdict as an approval log. (Rough edges: you
find the session by hand and there's no one-click dismiss — a dedicated
persistent approvals TUI + a gmux notify primitive are the follow-up.)
`HAKO_APPROVE_ALL` is a documented headless/CI escape hatch.

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
not of hako. The next refinement is a dedicated persistent approvals TUI (and a
gmux notify primitive) rather than one gmux session per request.
