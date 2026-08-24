// Command mcp is a Model Context Protocol server exposing sitemon tools over
// stdio, so any MCP client (e.g. Claude Code) can drive sitemon.
//
// Tools: check_url, ssl_check, and (when MONGO_URI is set) analyze.
package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mralaminahamed/sitemon/apps/ai/internal/analyzer"
	"github.com/mralaminahamed/sitemon/apps/ai/internal/llm"
	shttp "github.com/mralaminahamed/sitemon/packages/shared/http"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/mralaminahamed/sitemon/packages/shared/monitor"
	"github.com/mralaminahamed/sitemon/packages/shared/ssl"
	"github.com/mralaminahamed/sitemon/packages/shared/store"
)

type urlInput struct {
	URL string `json:"url"`
}

type analyzeInput struct {
	URL   string `json:"url"`
	Limit int    `json:"limit,omitempty"`
}

func main() {
	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{Name: "sitemon", Version: "0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_url",
		Description: "Run a live health check on a URL and return status, code, and latency.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in urlInput) (*mcp.CallToolResult, models.HealthResult, error) {
		client := shttp.NewClient(10 * time.Second)
		res, _ := monitor.NewHealthChecker(client).Check(normalize(in.URL))
		return nil, *res, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ssl_check",
		Description: "Inspect a URL's TLS certificate (issuer, expiry, days remaining).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in urlInput) (*mcp.CallToolResult, *ssl.CertificateInfo, error) {
		info, err := ssl.CheckCertificate(normalize(in.URL), 10*time.Second)
		return nil, info, err
	})

	if uri := os.Getenv("MONGO_URI"); uri != "" {
		checks, err := store.NewCheckStore(ctx, uri, envOr("MONGO_DB", "sitemon"))
		if err == nil {
			summarizer := llm.New()
			mcp.AddTool(server, &mcp.Tool{
				Name:        "analyze",
				Description: "Analyze stored check history for a URL: uptime, latency, anomalies, and an incident summary.",
			}, func(c context.Context, _ *mcp.CallToolRequest, in analyzeInput) (*mcp.CallToolResult, analyzer.Analysis, error) {
				limit := int64(in.Limit)
				if limit <= 0 {
					limit = 100
				}
				results, err := checks.History(c, normalize(in.URL), limit)
				if err != nil {
					return nil, analyzer.Analysis{}, err
				}
				a := analyzer.Analyze(normalize(in.URL), results)
				if s, err := summarizer.Summarize(c, a); err == nil {
					a.Summary = s
				}
				return nil, a, nil
			})
		}
	}

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, io.EOF) {
		log.Fatal(err)
	}
}

func normalize(u string) string {
	if u == "" {
		return u
	}
	if len(u) < 7 || (u[:7] != "http://" && (len(u) < 8 || u[:8] != "https://")) {
		return "https://" + u
	}
	return u
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
