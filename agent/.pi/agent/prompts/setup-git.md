Walk me through setting up git in this container. Read `docs/git.md` first; it is the source of truth, follow it.

1. **Identity.** Check `git config --global user.name` and `git config --global user.email`. If either is unset, ask me for the value and set it with `git config --global`. If both are already set, tell me what they are and move on.

2. **Push access.** Pushing needs a credential, and it is the one credential you are trusted to hold (see `docs/git.md` / ADR-0015). Check whether one is already configured (a `credential.helper`, or `~/.git-credentials`). If it is, confirm that and stop. If not, ask whether I want to set up push access now; if I say no, stop here.

   If yes, walk me through a fine-grained token:
   - Point me at <https://github.com/settings/personal-access-tokens/new>, "Only select repositories" (the repos you need), and **Repository permissions: Contents = Read and write, nothing else**.
   - Warn me to keep it Contents-only: a broader token would let you act on GitHub outside the approval gate, which defeats the point.
   - For storing it: you cannot type into git's interactive password prompt, so either tell me to run the store step in my own terminal (`hako shell`, then `git config --global credential.helper store` and a push), or -- only if I prefer -- I can hand you the token and you write it for me. Never ask me to paste the token into chat unless I choose that.

Finish with a one-line summary: identity set, and push configured or skipped. Do not set up anything I did not ask for.
