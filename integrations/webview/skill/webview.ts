#!/usr/bin/env bun
// webview -- fetch one or more URLs through the configured crawl4ai endpoint
// and print each page's extracted, boilerplate-pruned main content as markdown.
//
// Endpoint comes from HAKO_WEBVIEW_URL (set by the assembler from the
// integration's `url` setting); defaults to the bundled crawl4ai sidecar's DNS
// name. Holds no credentials -- fetching a public page is an unauthenticated read.
const base = (process.env.HAKO_WEBVIEW_URL ?? "http://webview:11235").replace(/\/+$/, "");
const MAX_CHARS = 10000; // ~2500 tokens/page; re-fetch with intent if you need more

const args = process.argv.slice(2);
const urls = args.filter((a) => a && !a.startsWith("-"));
if (urls.length === 0 || args.includes("-h") || args.includes("--help")) {
  console.log("usage: webview <url> [url ...]");
  process.exit(urls.length ? 0 : 2);
}
for (const u of urls) {
  if (!/^https?:\/\//.test(u)) {
    console.error(`webview: not a URL (need http/https): ${u}`);
    process.exit(2);
  }
}

// PruningContentFilter populates `fit_markdown` with the main content (nav,
// footers, cookie banners stripped) -- far fewer tokens than raw_markdown.
const crawler_config = {
  type: "CrawlerRunConfig",
  params: {
    markdown_generator: {
      type: "DefaultMarkdownGenerator",
      params: {
        content_filter: {
          type: "PruningContentFilter",
          params: { threshold: 0.48, threshold_type: "dynamic", min_word_threshold: 5 },
        },
      },
    },
  },
};

let res: Response;
try {
  res = await fetch(`${base}/crawl`, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify({ urls, crawler_config }),
    signal: AbortSignal.timeout(90_000),
  });
} catch {
  console.error(
    `webview: could not reach the crawl endpoint at ${base}. Is the webview ` +
      `sidecar up (hako up, with [integrations.webview] enabled), or is the url ` +
      `set? Agent: report this to the user.`,
  );
  process.exit(3);
}
if (!res.ok) {
  console.error(`webview: endpoint error HTTP ${res.status}`);
  process.exit(1);
}

const data: any = await res.json().catch(() => ({}));
const results: any[] = data.results ?? [];
if (results.length === 0) {
  console.error("webview: crawl4ai returned no results.");
  process.exit(1);
}

results.forEach((r, i) => {
  if (i > 0) console.log(`\n${"=".repeat(72)}\n`);
  const url = r.url ?? r.redirected_url ?? "";
  if (!r.success) {
    console.log(`# ${url}\n\n[fetch failed: ${r.error_message ?? "unknown error"}]`);
    return;
  }
  const md = r.markdown ?? {};
  let content = (typeof md === "string" ? md : md.fit_markdown || md.raw_markdown || "").trim();
  const title = (r.metadata?.title ?? "").trim();
  console.log(`# ${title || url}\n${url}\n`);
  if (!content) {
    console.log("[no extractable content]");
    return;
  }
  if (content.length > MAX_CHARS) {
    content =
      content.slice(0, MAX_CHARS) +
      `\n\n[truncated at ${MAX_CHARS} chars; full page was ${content.length}. ` +
      `Re-fetch only if you actually need more.]`;
  }
  console.log(content);
});
