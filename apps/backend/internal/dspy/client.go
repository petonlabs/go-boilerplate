package dspy

import (
	"context"
	"errors"
	"os"

	"github.com/XiaoConstantine/dspy-go/pkg/core"
	"github.com/XiaoConstantine/dspy-go/pkg/llms"
)

type Client struct {
	llm core.LLM
}

func New() (*Client, error) {
	if os.Getenv("DSPY_ENABLED") != "true" {
		return nil, errors.New("DSPY disabled")
	}
	if os.Getenv("DSPY_PROVIDER") != "azure" {
		return nil, errors.New("only azure provider supported in this setup")
	}
	endpoint := os.Getenv("DSPY_AZURE_ENDPOINT")
	key := os.Getenv("DSPY_AZURE_API_KEY")
	model := os.Getenv("DSPY_MODEL")
	if endpoint == "" || key == "" || model == "" {
		return nil, errors.New("missing DSPY envs")
	}

	// Configure Azure OpenAI via OpenAI-compatible API
	llm, err := llms.NewOpenAILLM(
		core.ModelOpenAIGPT4, // Model identifier
		llms.WithAPIKey(key),
		llms.WithOpenAIBaseURL(endpoint),
		llms.WithHeader("api-version", "2024-02-15-preview"),
	)
	if err != nil {
		return nil, err
	}

	return &Client{llm: llm}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	// Simple test to verify connectivity
	_, err := c.llm.Generate(ctx, "ping")
	return err
}
