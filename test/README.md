# test/

Testing fixtures and harnesses. Nothing here is needed to *run* hako — the repo
root stays limited to what a user needs to clone-and-up, and non-runtime
material (tests, fixtures, docs) lives in subfolders like this one.

Language-idiomatic unit tests stay beside their source (e.g. the launcher's
`launcher/*_test.go`, run with `go test ./...`); this folder is for cross-cutting
fixtures and harnesses that several pieces share.

- **`github-mock/`** — a tiny, fixture-backed MCP server mimicking the slice of
  the GitHub MCP that the `github` adapter calls. Lets the gateway + `callHook`
  loop be exercised end-to-end with the real CLI and zero credentials. See its
  `main.go` header for details; extend the fixtures as testing needs grow.
