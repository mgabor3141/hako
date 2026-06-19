#!/usr/bin/env bun
// websearch -- query the configured SearXNG endpoint and print results.
//
// Endpoint comes from HAKO_WEBSEARCH_URL (set by the assembler from the
// integration's `url` setting); defaults to the bundled sidecar's DNS name.
// Holds no credentials -- search is an unauthenticated read on the private net.
const base = (process.env.HAKO_WEBSEARCH_URL ?? "http://websearch:8080").replace(/\/+$/, "");
const MAX_RESULTS = 8; // hard cap: keep the result list cheap to read
const MAX_SNIPPET = 300; // SearXNG `content` can be 1-2k chars; 300 is plenty

const query = process.argv.slice(2).join(" ").trim();
if (!query || query === "--help" || query === "-h") {
  console.log("usage: websearch <query>");
  process.exit(query ? 0 : 2);
}

let res: Response;
try {
  res = await fetch(`${base}/search?q=${encodeURIComponent(query)}&format=json`);
} catch {
  console.error(
    `websearch: could not reach the search endpoint at ${base}. Is the websearch ` +
      `sidecar up (hako up), or is the url set? Agent: report this to the user.`,
  );
  process.exit(3);
}
if (!res.ok) {
  console.error(`websearch: endpoint error HTTP ${res.status}`);
  process.exit(1);
}

const data: any = await res.json().catch(() => ({}));
const results: any[] = (data.results ?? []).slice(0, MAX_RESULTS);
if (results.length === 0) {
  console.log("no results.");
} else {
  for (const r of results) {
    console.log(`${r.title ?? "(untitled)"}`);
    console.log(`  ${r.url ?? ""}`);
    const snippet = (r.content ?? "").trim().slice(0, MAX_SNIPPET);
    if (snippet) console.log(`  ${snippet}`);
  }
}
