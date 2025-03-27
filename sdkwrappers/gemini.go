package sdkwrappers

import (
	"context"
	"fmt"
	"time"

	"github.com/TrustSight-io/tokentracker/common"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Gemini model constants
const (
	GeminiPro    = "gemini-pro"
	GeminiUltra  = "gemini-ultra"
	GeminiPro1_5 = "gemini-1.5-pro"
	GeminiFlash  = "gemini-1.5-flash"
)

// GeminiSDKWrapper wraps the Gemini SDK client
type GeminiSDKWrapper struct {
	client *genai.Client
}

// NewGeminiSDKWrapper creates a new Gemini SDK wrapper
func NewGeminiSDKWrapper(apiKey string) (*GeminiSDKWrapper, error) {
	// Create client with API key
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiSDKWrapper{
		client: client,
	}, nil
}

// GetProviderName returns the name of the provider
func (w *GeminiSDKWrapper) GetProviderName() string {
	return "gemini"
}

// GetClient returns the underlying SDK client
func (w *GeminiSDKWrapper) GetClient() interface{} {
	return w.client
}

// GetSupportedModels returns a list of supported models
func (w *GeminiSDKWrapper) GetSupportedModels() ([]string, error) {
	// Hardcoded list of Gemini models
	return []string{
		GeminiPro,
		GeminiUltra,
		GeminiPro1_5,
		GeminiFlash,
	}, nil
}

// ExtractTokenUsageFromResponse extracts token usage from a Gemini API response
func (w *GeminiSDKWrapper) ExtractTokenUsageFromResponse(response interface{}) (common.TokenUsage, error) {
	// Try to cast the response to *genai.GenerateContentResponse
	resp, ok := response.(*genai.GenerateContentResponse)
	if !ok {
		return common.TokenUsage{}, fmt.Errorf("response is not a *genai.GenerateContentResponse: %T", response)
	}

	// Check if usage metadata is available
	if resp.UsageMetadata == nil {
		return common.TokenUsage{}, fmt.Errorf("response does not contain usage metadata")
	}

	// Extract token usage information
	usage := common.TokenUsage{
		InputTokens:    int(resp.UsageMetadata.PromptTokenCount),
		OutputTokens:   int(resp.UsageMetadata.CandidatesTokenCount),
		TotalTokens:    int(resp.UsageMetadata.TotalTokenCount),
		Timestamp:      time.Now(),
		PromptTokens:   int(resp.UsageMetadata.PromptTokenCount),
		ResponseTokens: int(resp.UsageMetadata.CandidatesTokenCount),
	}

	// Set model if available from candidates
	if len(resp.Candidates) > 0 && resp.Candidates[0] != nil {
		// The model information might be available in the candidate
		// This is a placeholder as the actual field depends on the SDK
		// usage.Model = resp.Candidates[0].ModelName
	}

	return usage, nil
}

// FetchCurrentPricing returns the current pricing for Gemini models
func (w *GeminiSDKWrapper) FetchCurrentPricing() (map[string]common.ModelPricing, error) {
	// Hardcoded pricing information for Gemini models
	// These values should be updated regularly or fetched from an API
	pricing := map[string]common.ModelPricing{
		GeminiPro: {
			InputPricePerToken:  0.00000025,
			OutputPricePerToken: 0.0000005,
			Currency:            "USD",
		},
		GeminiUltra: {
			InputPricePerToken:  0.00001,
			OutputPricePerToken: 0.00003,
			Currency:            "USD",
		},
		GeminiPro1_5: {
			InputPricePerToken:  0.0000005,
			OutputPricePerToken: 0.0000015,
			Currency:            "USD",
		},
		GeminiFlash: {
			InputPricePerToken:  0.00000025,
			OutputPricePerToken: 0.00000075,
			Currency:            "USD",
		},
	}

	return pricing, nil
}

// UpdateProviderPricing updates the pricing information in the provider
func (w *GeminiSDKWrapper) UpdateProviderPricing() error {
	// In a real implementation, this would update the pricing information in the provider
	// For now, we'll just return nil
	return nil
}

// TrackAPICall tracks an API call and returns usage metrics
func (w *GeminiSDKWrapper) TrackAPICall(model string, response interface{}) (common.UsageMetrics, error) {
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

// Close closes the client
func (w *GeminiSDKWrapper) Close() error {
	return w.client.Close()
}
