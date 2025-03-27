package sdkwrappers

import (
	"fmt"
	"time"

	"github.com/TrustSight-io/tokentracker/common"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Claude model constants
const (
	ClaudeHaiku  = "claude-3-haiku-20240307"
	ClaudeSonnet = "claude-3-sonnet-20240229"
	ClaudeOpus   = "claude-3-opus-20240229"
	ClaudeHaiku2 = "claude-3-haiku@20240307"
)

// AnthropicSDKWrapper wraps the Anthropic SDK client
type AnthropicSDKWrapper struct {
	client anthropic.Client
}

// NewAnthropicSDKWrapper creates a new Anthropic SDK wrapper
func NewAnthropicSDKWrapper(apiKey string) *AnthropicSDKWrapper {
	// Create client with API key
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &AnthropicSDKWrapper{
		client: client,
	}
}

// GetProviderName returns the name of the provider
func (w *AnthropicSDKWrapper) GetProviderName() string {
	return "anthropic"
}

// GetClient returns the underlying SDK client
func (w *AnthropicSDKWrapper) GetClient() interface{} {
	return w.client
}

// GetSupportedModels returns a list of supported models
func (w *AnthropicSDKWrapper) GetSupportedModels() ([]string, error) {
	// Hardcoded list of Claude models
	return []string{
		ClaudeHaiku,
		ClaudeSonnet,
		ClaudeOpus,
		ClaudeHaiku2,
	}, nil
}

// ExtractTokenUsageFromResponse extracts token usage from an Anthropic API response
func (w *AnthropicSDKWrapper) ExtractTokenUsageFromResponse(response interface{}) (common.TokenUsage, error) {
	// Try to cast the response to *anthropic.Message
	msg, ok := response.(*anthropic.Message)
	if !ok {
		return common.TokenUsage{}, fmt.Errorf("response is not an *anthropic.Message: %T", response)
	}

	// Extract token usage information
	usage := common.TokenUsage{
		InputTokens:    int(msg.Usage.InputTokens),
		OutputTokens:   int(msg.Usage.OutputTokens),
		TotalTokens:    int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
		CompletionID:   msg.ID,
		Model:          msg.Model,
		Timestamp:      time.Now(),
		PromptTokens:   int(msg.Usage.InputTokens),  // Same as InputTokens for Anthropic
		ResponseTokens: int(msg.Usage.OutputTokens), // Same as OutputTokens for Anthropic
	}

	return usage, nil
}

// FetchCurrentPricing returns the current pricing for Anthropic models
func (w *AnthropicSDKWrapper) FetchCurrentPricing() (map[string]common.ModelPricing, error) {
	// Hardcoded pricing information for Anthropic models
	// These values should be updated regularly or fetched from an API
	pricing := map[string]common.ModelPricing{
		ClaudeHaiku: {
			InputPricePerToken:  0.00000025,
			OutputPricePerToken: 0.00000125,
			Currency:            "USD",
		},
		ClaudeSonnet: {
			InputPricePerToken:  0.000003,
			OutputPricePerToken: 0.000015,
			Currency:            "USD",
		},
		ClaudeOpus: {
			InputPricePerToken:  0.00001,
			OutputPricePerToken: 0.00003,
			Currency:            "USD",
		},
		ClaudeHaiku2: {
			InputPricePerToken:  0.00000025,
			OutputPricePerToken: 0.00000125,
			Currency:            "USD",
		},
	}

	return pricing, nil
}

// UpdateProviderPricing updates the pricing information in the provider
func (w *AnthropicSDKWrapper) UpdateProviderPricing() error {
	// In a real implementation, this would update the pricing information in the provider
	// For now, we'll just return nil
	return nil
}

// TrackAPICall tracks an API call and returns usage metrics
func (w *AnthropicSDKWrapper) TrackAPICall(model string, response interface{}) (common.UsageMetrics, error) {
	// Extract token usage from the response
	tokenUsage, err := w.ExtractTokenUsageFromResponse(response)
	if err != nil {
		return common.UsageMetrics{}, err
	}

	// Get pricing information for the model
	pricing, err := w.FetchCurrentPricing()
	if err != nil {
		return common.UsageMetrics{}, err
	}

	modelPricing, ok := pricing[model]
	if !ok {
		return common.UsageMetrics{}, fmt.Errorf("no pricing information found for model: %s", model)
	}

	// Calculate price
	inputCost := float64(tokenUsage.InputTokens) * modelPricing.InputPricePerToken
	outputCost := float64(tokenUsage.OutputTokens) * modelPricing.OutputPricePerToken
	totalCost := inputCost + outputCost

	// Create usage metrics
	metrics := common.UsageMetrics{
		TokenCount: common.TokenCount{
			InputTokens:    tokenUsage.InputTokens,
			ResponseTokens: tokenUsage.OutputTokens,
			TotalTokens:    tokenUsage.TotalTokens,
		},
		Price: common.Price{
			InputCost:  inputCost,
			OutputCost: outputCost,
			TotalCost:  totalCost,
			Currency:   modelPricing.Currency,
		},
		Duration:  time.Since(tokenUsage.Timestamp),
		Timestamp: time.Now(),
		Model:     model,
		Provider:  w.GetProviderName(),
	}

	return metrics, nil
}
