# test/

Testing fixtures and harnesses. Nothing here is needed to *run* hako — the repo
root stays limited to what a user needs to clone-and-up, and non-runtime
material (tests, fixtures, docs) lives in subfolders like this one.

- **`github-mock/`** — a tiny, fixture-backed MCP server mimicking the slice of
  the GitHub MCP that the `github` adapter calls. Lets the broker + `callHook`
  loop be exercised end-to-end with the real CLI and zero credentials. See its
  `main.go` header for details; extend the fixtures as testing needs grow.
