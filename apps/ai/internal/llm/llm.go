// Package llm turns an analysis into a short incident summary via Claude.
package llm

import (
	"context"
	"encoding/json"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mralaminahamed/sitemon/apps/ai/internal/analyzer"
)

type Summarizer interface {
	Summarize(ctx context.Context, a analyzer.Analysis) (string, error)
}

// Noop is used when no API key is configured.
type Noop struct{}

func (Noop) Summarize(context.Context, analyzer.Analysis) (string, error) { return "", nil }

type Claude struct {
	client anthropic.Client
	model  string
}

// New returns a Claude summarizer when ANTHROPIC_API_KEY is set, otherwise Noop.
func New() Summarizer {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return Noop{}
	}
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-opus-5"
	}
	return &Claude{client: anthropic.NewClient(), model: model}
}

const system = "You are an SRE assistant. Given JSON uptime and latency metrics for a monitored URL, write a 2-3 sentence incident summary: current state, the most likely cause, and one concrete recommendation. Be concise and specific. Do not restate every number."

func (c *Claude) Summarize(ctx context.Context, a analyzer.Analysis) (string, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 400,
		System:    []anthropic.TextBlockParam{{Text: system}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(string(data))),
		},
	})
	if err != nil {
		return "", err
	}
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			return t.Text, nil
		}
	}
	return "", nil
}
