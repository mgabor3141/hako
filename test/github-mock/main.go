// github-mock is a tiny, fixture-backed MCP server that mimics the slice of the
// GitHub MCP which the `github` CLI (agent/.agents/skills/github) calls. It lets
// the broker + callHook loop be exercised end-to-end with the real CLI and zero
// credentials. It is NOT faithful GitHub behavior — just enough response shape
// for the CLI's parsers. Extend the fixtures/tools as testing needs grow.
//
// Speaks streamable HTTP via mcp-go (the same library the broker uses, so the
// broker can proxy it). Listens on :8080 at endpoint /mcp by default; override
// with ADDR. Run standalone: `go run .`
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type obj = map[string]any

// recent is a timestamp the CLI's relative-time formatter ("2h ago") can render.
var recent = time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)

func result(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	// The CLI does JSON.parse(content[0].text); a single text part is enough.
	return mcp.NewToolResultText(string(b)), nil
}

// method reads the "method" argument used by the CLI's multiplexed tools
// (pull_request_read, issue_read/write, actions_*).
func method(req mcp.CallToolRequest) string {
	if m := req.GetArguments(); m != nil {
		if s, ok := m["method"].(string); ok {
			return s
		}
	}
	return ""
}

func main() {
	s := server.NewMCPServer("github-mock", "0.0.1")

	user := obj{"login": "octomock"}
	pr := obj{
		"number": 1, "title": "Mock PR", "state": "open", "merged": false, "draft": false,
		"user": user, "base": obj{"ref": "main"}, "head": obj{"ref": "feature"},
		"additions": 3, "deletions": 1, "changed_files": 2,
		"created_at": recent, "updated_at": recent, "body": "Mock pull request body.",
	}
	issue := obj{
		"number": 1, "title": "Mock issue", "state": "open", "user": user,
		"labels": []any{}, "created_at": recent, "updated_at": recent, "body": "Mock issue body.",
	}

	add := func(name string, h server.ToolHandlerFunc) {
		s.AddTool(mcp.NewTool(name, mcp.WithDescription("github-mock: "+name)), h)
	}

	// ── reads (ungated) ──────────────────────────────────────────────────────
	add("get_me", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result(user)
	})
	add("list_pull_requests", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result([]any{pr})
	})
	add("pull_request_read", func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		switch method(req) {
		case "get_check_runs":
			return result(obj{"check_runs": []any{obj{"name": "ci", "status": "completed", "conclusion": "success", "id": 1}}})
		case "get_reviews":
			return result(obj{"reviews": []any{}})
		case "get_review_comments":
			return result(obj{"review_threads": []any{}})
		case "get_comments":
			return result([]any{})
		default: // "get"
			return result(pr)
		}
	})
	add("get_commit", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result(obj{"commit": obj{"committer": obj{"date": recent}}})
	})
	add("issue_read", func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if method(req) == "get_comments" {
			return result([]any{})
		}
		return result(issue)
	})
	add("list_issues", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result(obj{"issues": []any{issue}})
	})

	// ── writes (the tools worth gating with callHook) ────────────────────────
	add("create_pull_request", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result(obj{"number": 42, "html_url": "https://github.com/octomock/mock/pull/42"})
	})
	add("update_pull_request", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result(obj{"number": 1})
	})
	add("merge_pull_request", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result(obj{"merged": true})
	})
	add("add_issue_comment", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result(obj{"id": 100})
	})
	add("issue_write", func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if method(req) == "create" {
			return result(obj{"number": 7, "html_url": "https://github.com/octomock/mock/issues/7"})
		}
		return result(obj{"number": 1})
	})
	add("pull_request_review_write", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return result(obj{"ok": true})
	})

	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	log.Printf("github-mock listening on %s (endpoint /mcp)", addr)
	if err := server.NewStreamableHTTPServer(s, server.WithStateLess(true)).Start(addr); err != nil {
		log.Fatal(err)
	}
}
