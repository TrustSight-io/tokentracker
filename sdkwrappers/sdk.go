// Package sdkwrappers provides adapters for official LLM SDK clients
package sdkwrappers

import (
	"github.com/TrustSight-io/tokentracker/common"
)

// SDKClientWrapper defines the interface for wrapping official LLM SDK clients
type SDKClientWrapper interface {
	// GetProviderName returns the name of the LLM provider (e.g., "openai", "anthropic", "gemini")
	GetProviderName() string

	// GetClient returns the underlying SDK client instance
	GetClient() interface{}

	// GetSupportedModels returns a list of model identifiers supported by this provider
	GetSupportedModels() ([]string, error)

	// ExtractTokenUsageFromResponse extracts token usage information from an API response
	ExtractTokenUsageFromResponse(response interface{}) (common.TokenUsage, error)

	// FetchCurrentPricing fetches the current pricing information for all supported models
	FetchCurrentPricing() (map[string]common.ModelPricing, error)
	
	// UpdateProviderPricing updates the pricing information in the provider
	UpdateProviderPricing() error
	
	// TrackAPICall tracks an API call and returns usage metrics
	TrackAPICall(model string, response interface{}) (common.UsageMetrics, error)
}
