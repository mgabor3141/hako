---
name: websearch
description: >
  Search the web via the `websearch` CLI. Use when a task needs current
  information, documentation, or facts not in the repo or your context.
---

# websearch

Run `websearch <query>` to search the web. Prints a short list of results
(title, URL, snippet).

```sh
websearch how to pin a docker image by digest
websearch "site:docs.docker.com compose env_file required"
```

Notes:
- One query per call; refine and re-run rather than paging. Returns up to 8
  results, snippets capped short -- enough to decide what to open.
- It hits whatever endpoint hako wired (a bundled SearXNG sidecar by default,
  or a configured `url`). It holds no credentials.
- To read a specific page's contents, use the `webview` skill (if enabled) or
  fetch/clone it directly -- don't search for a URL you already have.
