Walk me through setting up git in this container.

1. **Identity.** Check `git config --global user.name` and `git config --global user.email`. If either is unset, ask me for the value and set it with `git config --global`.

2. **Push access.** Pushing needs a credential, and it is the one credential you are trusted to hold. Check whether one is already configured (a `credential.helper`, or `~/.git-credentials`).

   If it isn't, walk me through a fine-grained token:
   - Point me at <https://github.com/settings/personal-access-tokens/new>, "Only select repositories" (the repos you need), and **Repository permissions: Contents = Read and write, nothing else**.
   - Warn me to keep it Contents-only: a broader token would let you act on GitHub outside the approval gate, which defeats the point.
   - Ask me to paste, then store it in the appropriate repo-level or global git config.
