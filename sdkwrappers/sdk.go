// Package sdkwrappers provides adapters for official LLM SDK clients
package sdkwrappers

import (
	"time"

	"github.com/TrustSight-io/tokentracker"
)

// TokenUsage represents token usage information extracted from API responses
type TokenUsage struct {
	InputTokens    int
	OutputTokens   int
	TotalTokens    int
	CompletionID   string
	Model          string
	Timestamp      time.Time
	PromptTokens   int    // Some APIs use "prompt" instead of "input"
	ResponseTokens int    // Some APIs use "response" instead of "output"
	RequestID      string // Some APIs provide a request ID
}

// SDKClientWrapper defines the interface for wrapping official LLM SDK clients
type SDKClientWrapper interface {
	// GetProviderName returns the name of the LLM provider (e.g., "openai", "anthropic", "gemini")
	GetProviderName() string

	// GetClient returns the underlying SDK client instance
	GetClient() interface{}

	// GetSupportedModels returns a list of model identifiers supported by this provider
	GetSupportedModels() ([]string, error)

	// ExtractTokenUsageFromResponse extracts token usage information from an API response
	ExtractTokenUsageFromResponse(response interface{}) (TokenUsage, error)

	// FetchCurrentPricing fetches the current pricing information for all supported models
	FetchCurrentPricing() (map[string]tokentracker.ModelPricing, error)
}