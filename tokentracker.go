package tokentracker

import (
	"fmt"
	"time"

	"github.com/TrustSight-io/tokentracker/sdkwrappers"
)

// TokenTracker interface defines the main functionality
type TokenTracker interface {
	// CountTokens counts tokens for a text string or chat messages
	CountTokens(params TokenCountParams) (TokenCount, error)

	// CalculatePrice calculates price based on token usage
	CalculatePrice(model string, inputTokens, outputTokens int) (Price, error)

	// TrackUsage tracks full usage for an LLM call
	TrackUsage(callParams CallParams, response interface{}) (UsageMetrics, error)

	// RegisterSDKClient registers an SDK client with the appropriate provider
	RegisterSDKClient(client sdkwrappers.SDKClientWrapper) error

	// UpdateAllPricing updates pricing information for all registered providers
	UpdateAllPricing() error

	// TrackTokenUsage extracts token usage from a provider response
	TrackTokenUsage(providerName string, response interface{}) (TokenCount, error)
}

// DefaultTokenTracker implements the TokenTracker interface
type DefaultTokenTracker struct {
	registry *ProviderRegistry
	config   *Config
}

// NewTokenTracker creates a new token tracker with the given configuration
func NewTokenTracker(config *Config) *DefaultTokenTracker {
	registry := NewProviderRegistry()

	// Register default providers here or allow caller to register them

	return &DefaultTokenTracker{
		registry: registry,
		config:   config,
	}
}

// RegisterProvider registers a provider with the token tracker
func (t *DefaultTokenTracker) RegisterProvider(provider Provider) {
	t.registry.Register(provider)
}

// RegisterSDKClient registers an SDK client with the appropriate provider
func (t *DefaultTokenTracker) RegisterSDKClient(client sdkwrappers.SDKClientWrapper) error {
	providerName := client.GetProviderName()
	provider, exists := t.registry.Get(providerName)
	
	if !exists {
		return NewError(ErrProviderNotFound, fmt.Sprintf("no provider found with name: %s", providerName), nil)
	}
	
	// Set the SDK client in the provider
	provider.SetSDKClient(client.GetClient())
	
	// Update pricing information
	if err := client.UpdateProviderPricing(); err != nil {
		return NewError(ErrPricingUpdateFailed, "failed to update pricing information", err)
	}
	
	return nil
}

// UpdateAllPricing updates pricing information for all registered providers
func (t *DefaultTokenTracker) UpdateAllPricing() error {
	providers := t.registry.All()
	var lastErr error
	
	for _, provider := range providers {
		if err := provider.UpdatePricing(); err != nil {
			lastErr = err
		}
	}
	
	if lastErr != nil {
		return NewError(ErrPricingUpdateFailed, "failed to update pricing for one or more providers", lastErr)
	}
	
	return nil
}

// TrackTokenUsage extracts token usage from a provider response
func (t *DefaultTokenTracker) TrackTokenUsage(providerName string, response interface{}) (TokenCount, error) {
	provider, exists := t.registry.Get(providerName)
	
	if !exists {
		return TokenCount{}, NewError(ErrProviderNotFound, fmt.Sprintf("no provider found with name: %s", providerName), nil)
	}
	
	return provider.ExtractTokenUsageFromResponse(response)
}

// CountTokens counts tokens for the given parameters
func (t *DefaultTokenTracker) CountTokens(params TokenCountParams) (TokenCount, error) {
	if params.Model == "" {
		return TokenCount{}, NewError(ErrInvalidParams, "model is required", nil)
	}

	provider, exists := t.registry.GetForModel(params.Model)
	if !exists {
		return TokenCount{}, NewError(ErrProviderNotFound, fmt.Sprintf("no provider found for model: %s", params.Model), nil)
	}

	return provider.CountTokens(params)
}

// CalculatePrice calculates price based on token usage
func (t *DefaultTokenTracker) CalculatePrice(model string, inputTokens, outputTokens int) (Price, error) {
	if model == "" {
		return Price{}, NewError(ErrInvalidParams, "model is required", nil)
	}

	provider, exists := t.registry.GetForModel(model)
	if !exists {
		return Price{}, NewError(ErrProviderNotFound, fmt.Sprintf("no provider found for model: %s", model), nil)
	}

	return provider.CalculatePrice(model, inputTokens, outputTokens)
}

// TrackUsage tracks full usage for an LLM call
func (t *DefaultTokenTracker) TrackUsage(callParams CallParams, response interface{}) (UsageMetrics, error) {
	// Get input token count
	inputCount, err := t.CountTokens(callParams.Params)
	if err != nil {
		return UsageMetrics{}, err
	}

	// Extract response tokens from the response object
	// This will be provider-specific and depend on the response structure
	var outputTokens int

	// Try to extract token count from response if it's available
	if extractor, ok := response.(interface {
		GetTokenCount() int
	}); ok {
		outputTokens = extractor.GetTokenCount()
	} else {
		// Fallback to estimating response tokens
		provider, exists := t.registry.GetForModel(callParams.Model)
		if exists {
			// Create a new params object with CountResponseTokens set to true
			estimateParams := callParams.Params
			estimateParams.CountResponseTokens = true
			estimate, err := provider.CountTokens(estimateParams)
			if err == nil {
				outputTokens = estimate.ResponseTokens
			}
		}
	}

	// Calculate price
	price, err := t.CalculatePrice(callParams.Model, inputCount.InputTokens, outputTokens)
	if err != nil {
		return UsageMetrics{}, err
	}

	// Calculate duration
	duration := time.Since(callParams.StartTime)

	// Get provider name
	provider, _ := t.registry.GetForModel(callParams.Model)
	providerName := provider.Name()

	// Create usage metrics
	metrics := UsageMetrics{
		TokenCount: TokenCount{
			InputTokens:    inputCount.InputTokens,
			ResponseTokens: outputTokens,
			TotalTokens:    inputCount.InputTokens + outputTokens,
		},
		Price:     price,
		Duration:  duration,
		Timestamp: time.Now(),
		Model:     callParams.Model,
		Provider:  providerName,
	}

	return metrics, nil
}
