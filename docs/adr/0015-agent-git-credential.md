# ADR-0015: The agent may hold a scoped git push credential

- **Status:** Accepted -- 2026-06-19

## Context
Real work means opening PRs, which means `git push`. git runs locally in the
container and authenticates locally, so there is no way to push without a
credential present in the container. That collides with ADR-0002 (the agent
holds no credentials; the boundary is the *absence* of secrets).

We looked for ways to keep the rule intact:
- **Commit via the gateway** (the GitHub MCP file/branch tools). Rejected: it
  makes the agent jump through hoops for something it already knows how to do.
  Pure git should stay native git; only GitHub-API actions belong in the gateway.
- **A git smart-HTTP proxy** in the gateway that injects the token server-side
  and gates pushes. The pure answer (token never enters the container, push
  stays approval-gated), but a whole new subsystem against the "fewer moving
  parts" rule, and its main benefit is a property we choose not to need.
- **SSH deploy keys** are the tightest credential (one repo, transport-only, no
  API), but adding one needs repo *admin*, so they do not work for repos you
  only have write on.
- **Agent self-provisioning via a gated MCP tool.** Tempting, and the per-repo
  grant *is* a genuine, deny-able authorization decision worth having. But
  minting a *persistent* credential has exactly one mechanism -- a deploy key via
  `POST /repos/{owner}/{repo}/keys` -- which needs an **Administration**-scoped
  token in the gateway. PATs cannot be API-minted at all, and GitHub App
  installation tokens are ephemeral (1h), not the lasting grant we want. An
  Administration token in the gateway is a far worse thing to leak than a
  Contents credential held by the agent (see the `/proc`-readable-token gap in
  production-readiness.md), so auto-provisioning is rejected. Granting push
  access stays a human action: the agent asks, the human sets up the credential.

## Decision
Pushing is the one bounded exception to ADR-0002. The agent may hold a
**transport-only** git credential:

- A human provisions it with **stock git inside the persistent home**
  ([`docs/git.md`](../git.md)); hako ships **guidance, not tooling** -- no
  `hako git` command, no credential helper, no key management. git already
  handles global/per-repo/per-path credentials, and the bind-mounted home makes
  the setup persist.
- Recommended: a fine-grained PAT with **Contents: write** only, scoped to the
  chosen repos (a deploy key where you have admin; fork-and-PR for upstreams you
  do not).
- It stays **separate from and narrower than** the gateway's github token.
  PR/issue/CI writes remain on the gateway token, approval-gated; the push
  credential cannot open or merge PRs.
- `main` is protected on GitHub (no force-push, no direct push, require a
  reviewed PR). A Contents token cannot override a ruleset, so landing changes
  still flows through the gated PR path.
- **No token-scope validation.** GitHub exposes no introspection for a
  fine-grained token's permissions, so we communicate the scoping clearly and
  trust the operator instead of building a flaky check.

## Why
git transport genuinely cannot work without a local credential, and wrapping it
or proxying it costs more than the rule is worth here. Scoping the exception
tightly -- transport-only, repos you opt into, PR landing still gated, `main`
protected -- bounds the blast radius to "unapproved pushes to feature branches
of repos you chose," which is recoverable and is the agent's job anyway. The
credential lives with the agent because git must read it; the vault (which the
agent must *not* read) is the wrong home for it.

## Consequences
- The agent can `git push` to repos whose credential you set up; nothing else
  about the model changes. Opening/merging PRs still goes through the gated
  gateway token.
- Any code in the container can use that credential to push to those repos
  without approval. That is the accepted trade: keep the token Contents-only and
  the repo set small.
- On a missing credential the agent asks you to set one up; it never
  self-provisions, so granting push access stays a human action.
- The credential sits plaintext in the gitignored home (uncommittable).
  Encrypting it at rest is possible later, but the agent reads it at runtime
  regardless, so the gain would be small.
