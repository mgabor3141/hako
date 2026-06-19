---
name: webview
description: >
  Read a web page's main content as markdown via the `webview` CLI. Use when you
  have a URL (often from `websearch`) and need what the page actually says --
  docs, an issue, an article -- without the nav/ads/boilerplate.
---

# webview

Run `webview <url>` to fetch a page and print its extracted main content as
markdown (boilerplate stripped). Pass several URLs to fetch them in one batch.

```sh
webview https://docs.docker.com/compose/compose-file/
webview https://github.com/owner/repo/issues/42 https://example.com/post
```

Notes:
- Content is pruned to the main article and truncated (~10k chars/page) to keep
  your context window sane -- re-fetch a single URL only if you truly need more.
- It hits whatever endpoint hako wired (a bundled crawl4ai sidecar by default,
  or a configured `url`). It holds no credentials, so it only reaches pages that
  need none.
- Fetch sparingly: read `websearch` snippets first, then `webview` the one or
  two pages that actually matter.
