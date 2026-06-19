# Giving the agent git access

The agent works in real git checkouts in its home. Cloning and reading public
repos needs nothing. **Pushing** is the only part that needs a credential, and
it is the one deliberate exception to hako's "agent holds no credentials" rule
(ADR-0002; see [ADR-0015](./adr/0015-agent-git-credential.md)): git runs inside
the container and authenticates locally, so the push credential lives *with* the
agent. Scope it tightly.

You set this up with stock git inside the container, and it persists: the home
is bind-mounted, and `agent/` ignores everything except shipped config, so your
identity, `~/.git-credentials`, and `~/.ssh/` are never committed.

## Identity

```sh
hako shell      # or: docker exec -it hako zsh
git config --global user.name  "Your Name"
git config --global user.email "you@example.com"
```

## A push credential (recommended: a scoped fine-grained token)

Create one at <https://github.com/settings/personal-access-tokens/new>:

- **Only select repositories** -> the repos the agent may push to.
- **Repository permissions: Contents -> Read and write**, nothing else. (Add
  **Workflows -> Read and write** only if the agent will edit
  `.github/workflows/**`.)

Then store it inside the container:

```sh
git config --global credential.helper store
git push        # prompts once: username = anything, password = the token
                # saved to ~/.git-credentials (gitignored, survives restarts)
```

Different tokens for different repos or owners? Add
`git config --global credential.useHttpPath true` and store one per path. A
single fine-grained token already covers every repo you selected under its
owner, so most people set one and forget it.

## Two things that keep this safe

- **Keep this token Contents-only.** If it also had Pull requests or Issues
  write, the agent could open, merge, and comment via the API directly,
  bypassing the gateway's approval gate. Those actions stay on the *gateway*
  token (`hako auth github`), which is gated. This token only moves commits.
- **Protect `main` on GitHub** (branch protection or a ruleset: no force-push,
  no direct push, require a PR). A Contents token cannot override a ruleset, so
  the agent pushes feature branches freely but can only *land* changes through a
  reviewed PR, which still flows through the gated gateway token.

## How the agent uses it

It just runs `git`. If a push fails because no credential is set up for that
repo, it stops and asks you to set one up (pointing here). It never invents or
fetches a token on its own; granting push access is always a human action.

## Tighter or admin-free alternatives

- **SSH deploy key** -- tightest (one repo, transport-only, no API at all), but
  adding one needs *admin* on the repo, so it only works for repos you own or
  admin (including your own forks). Generate a key in `~/.ssh/`, add the public
  half to the repo's Settings -> Deploy keys with write access.
- **Fork-and-PR** -- for upstreams you do not admin: the agent pushes to your
  fork (which you do admin) and opens a cross-repo PR through the `github` tool.
