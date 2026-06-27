# ADR-0016: Tool-call approval is a liveness-gated gmux notification

- **Status:** Proposed — 2026-06-27. **Refines ADR-0010** (tool-call approval is
  a swappable hook): this names the *default* hook implementation and the surface
  it uses. Depends on gmux ADR 0018 (the notification primitive).

## Context

ADR-0010 made tool-call approval a swappable `callHook`, with the initial
gmux-session prompt explicitly a **stopgap UI**. We want a better default surface
that (a) reaches the user wherever they are, **including offline**, and (b) does
**not fail open**.

Two roads were explored and rejected:

- **executor.sh as the gateway/approval engine** (issue #4): its async
  pause/resume and runtime-DB connection model don't fit hako's bash-native,
  file-first, declarative design, and its credential/approval wins are largely
  things hako already has (the agent holds no credentials; the vault encrypts at
  rest; `callHook` already blocks *transparently*, which executor's CLI does not).
- **A forwarded web approval UI in gmux:** `ts.net` is on the Public Suffix List,
  so every tailnet node is *same-site* with gmux — a separate forwarding origin
  buys almost no isolation and adds a web surface to an RCE-equivalent control
  plane (`/new`). Rejected.

Meanwhile gmux is building a **stateful notification primitive** (gmux ADR 0018):
interactive, resolve-once actions; presence-routed delivery (in-app, with silent/
alerting mirrors to external channels like Telegram); a history view; and a
**liveness-bound await** connection. hako should consume it rather than grow its
own UI.

## Decision

**hako's default approval hook registers the gated call as an interactive gmux
notification and blocks on its resolution.**

- The hook (in the gateway) opens a **long-lived register+await connection to
  gmux**: *"Tool X wants to do Y — [Approve] [Deny]"*, carrying the originating
  gmux session id (propagated by the MCP CLI adapters) as provenance.
- **hako owns the gate; gmux owns the notification.** hako translates the
  resolution into the blocked call's outcome:
  - resolved `approve` ⇒ **allow**; `deny` ⇒ **deny**;
  - resolved `withdrawn` / `expired` ⇒ **fail closed** (call aborted/denied).
- **Liveness is the primary gate.** The await-connection's liveness is the
  validity condition. If the agent dies or the host reboots, the connection
  drops, gmux withdraws the notification, and the call fails closed — we never
  approve an operation that no agent is waiting to receive. The connection is
  held **locally** (gateway ↔ gmux), so an indefinite wait is fine; a network
  blip is tolerated by **reconnecting with the same notification id**.
- **Optional per-tool max-wait, off by default.** Liveness — not a timer — is the
  default gate. When a max-wait is set and elapses, the result is `expired` and
  the call fails closed.
- **Adapter scaffolding:** hako's MCP CLI adapters **propagate the gmux session
  id** (env var) on every gated call, so approvals are attributable to a session.
  This is a provenance hint only, **not a trust signal** (the gmux side is already
  `/new`-level authed; spoofing the id gains nothing).

## Consequences

- The stopgap gmux-session prompt is replaced; approvals reach the user **in-app
  or via gmux's external channels (Telegram, …), offline included**, without hako
  knowing or caring which channel.
- **Fail-closed is structural:** absence of a live caller ⇒ `withdrawn` ⇒ denied.
  No TTL is needed for safety; a max-wait is optional defense-in-depth.
- hako holds **no web surface and no notification UI of its own**; it depends only
  on gmux ADR 0018's contract — the single cross-project dependency.
- The hook stays **swappable** (ADR-0010): a non-gmux deployment can supply a
  different hook. The gmux notification hook is the *default*, not the only option.
- The latency model is unified: the **long-lived connection is primary** (and
  encodes liveness); "call/response" is just reconnection for blips — there is no
  second protocol to maintain.

## Alternatives Considered

- **executor.sh as gateway/approval engine** — rejected (issue #4).
- **Forwarded web approval UI** — rejected (`ts.net` PSL ⇒ same-site; RCE control
  plane).
- **Pure async pause/resume** (decouple approval from a live caller) — rejected as
  the default: it permits approving operations with no agent waiting. Liveness
  binding is the desired safety property. Could be offered for genuinely long
  offline waits later, behind explicit policy.
- **Mandatory wall-clock TTL as the primary gate** — rejected: liveness is a
  better primary gate; a TTL is optional.
