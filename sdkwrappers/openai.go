package sdkwrappers

import (
	"fmt"
	"time"

	"github.com/TrustSight-io/tokentracker/common"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAI model constants
const (
	GPT35Turbo    = "gpt-3.5-turbo"
	GPT35Turbo16K = "gpt-3.5-turbo-16k"
	GPT4          = "gpt-4"
	GPT4Turbo     = "gpt-4-turbo"
	GPT4o         = "gpt-4o"
)

// OpenAISDKWrapper wraps the OpenAI SDK client
type OpenAISDKWrapper struct {
	client openai.Client
}

// NewOpenAISDKWrapper creates a new OpenAI SDK wrapper
func NewOpenAISDKWrapper(apiKey string) *OpenAISDKWrapper {
	// Create client with API key
	client := openai.NewClient(option.WithAPIKey(apiKey))
	
	return &OpenAISDKWrapper{
		client: client,
	}
}

// GetProviderName returns the name of the provider
func (w *OpenAISDKWrapper) GetProviderName() string {
	return "openai"
}

// GetClient returns the underlying SDK client
func (w *OpenAISDKWrapper) GetClient() interface{} {
	return w.client
}

// GetSupportedModels returns a list of supported models
func (w *OpenAISDKWrapper) GetSupportedModels() ([]string, error) {
	// Hardcoded list of OpenAI models
	return []string{
		GPT35Turbo,
		GPT35Turbo16K,
		GPT4,
		GPT4Turbo,
		GPT4o,
	}, nil
}

// ExtractTokenUsageFromResponse extracts token usage from an OpenAI API response
func (w *OpenAISDKWrapper) ExtractTokenUsageFromResponse(response interface{}) (common.TokenUsage, error) {
	// Try to cast the response to *openai.ChatCompletion
	resp, ok := response.(*openai.ChatCompletion)
	if !ok {
		return common.TokenUsage{}, fmt.Errorf("response is not a *openai.ChatCompletion: %T", response)
	}

	// Extract token usage information
	usage := common.TokenUsage{
		InputTokens:    int(resp.Usage.PromptTokens),
		OutputTokens:   int(resp.Usage.CompletionTokens),
		TotalTokens:    int(resp.Usage.TotalTokens),
		CompletionID:   resp.ID,
		Model:          resp.Model,
		Timestamp:      time.Now(),
		PromptTokens:   int(resp.Usage.PromptTokens),    // Same as InputTokens for OpenAI
		ResponseTokens: int(resp.Usage.CompletionTokens), // Same as OutputTokens for OpenAI
		RequestID:      resp.SystemFingerprint,          // OpenAI uses SystemFingerprint as a request ID
	}

	return usage, nil
}

// FetchCurrentPricing returns the current pricing for OpenAI models
func (w *OpenAISDKWrapper) FetchCurrentPricing() (map[string]common.ModelPricing, error) {
	// Hardcoded pricing information for OpenAI models
	// These values should be updated regularly or fetched from an API
	pricing := map[string]common.ModelPricing{
		GPT35Turbo: {
			InputPricePerToken:  0.0000015,
			OutputPricePerToken: 0.000002,
			Currency:            "USD",
		},
		GPT35Turbo16K: {
			InputPricePerToken:  0.000003,
			OutputPricePerToken: 0.000004,
			Currency:            "USD",
		},
		GPT4: {
			InputPricePerToken:  0.00003,
			OutputPricePerToken: 0.00006,
			Currency:            "USD",
		},
		GPT4Turbo: {
			InputPricePerToken:  0.00001,
			OutputPricePerToken: 0.00003,
			Currency:            "USD",
		},
		GPT4o: {
			InputPricePerToken:  0.00001,
			OutputPricePerToken: 0.00003,
			Currency:            "USD",
		},
	}

	return pricing, nil
}

// UpdateProviderPricing updates the pricing information in the provider
func (w *OpenAISDKWrapper) UpdateProviderPricing() error {
	// In a real implementation, this would update the pricing information in the provider
	// For now, we'll just return nil
	return nil
}

// TrackAPICall tracks an API call and returns usage metrics
func (w *OpenAISDKWrapper) TrackAPICall(model string, response interface{}) (common.UsageMetrics, error) {
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

	// Use the model parameter instead of extracting from response
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