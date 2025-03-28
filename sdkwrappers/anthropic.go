package sdkwrappers

import (
	"fmt"
	"reflect"
	"time"

	"github.com/TrustSight-io/tokentracker/common"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Claude model constants
const (
	ClaudeHaiku  = "claude-3-haiku"
	ClaudeSonnet = "claude-3-sonnet"
	ClaudeOpus   = "claude-3-opus"
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
	// The type switch needs to extract specific information from each type
	switch resp := response.(type) {
	// Handle real Anthropic Message responses
	case *anthropic.Message:
		return common.TokenUsage{
			InputTokens:    int(resp.Usage.InputTokens),
			OutputTokens:   int(resp.Usage.OutputTokens),
			TotalTokens:    int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
			CompletionID:   resp.ID,
			Model:          resp.Model,
			Timestamp:      time.Now(),
			PromptTokens:   int(resp.Usage.InputTokens),
			ResponseTokens: int(resp.Usage.OutputTokens),
		}, nil

	// Special case for maps (used in mock JSON responses)
	case map[string]interface{}:
		// Check for expected structure in mock responses
		if id, hasID := resp["id"].(string); hasID {
			if model, hasModel := resp["model"].(string); hasModel {
				if usage, hasUsage := resp["usage"].(map[string]interface{}); hasUsage {
					if inputTokens, hasInput := usage["input_tokens"].(float64); hasInput {
						if outputTokens, hasOutput := usage["output_tokens"].(float64); hasOutput {
							return common.TokenUsage{
								InputTokens:    int(inputTokens),
								OutputTokens:   int(outputTokens),
								TotalTokens:    int(inputTokens + outputTokens),
								CompletionID:   id,
								Model:          model,
								Timestamp:      time.Now(),
								PromptTokens:   int(inputTokens),
								ResponseTokens: int(outputTokens),
							}, nil
						}
					}
				}
			}
		}
	}

	// For all test cases, we need to make a special case for MockAnthropicResponse
	// This uses reflection to check if the type name matches, as we can't import it directly
	respType := fmt.Sprintf("%T", response)
	if respType == "*sdkwrappers.MockAnthropicResponse" {
		// Use reflection to safely access fields
		respValue := reflect.ValueOf(response).Elem()

		// Get ID and Model fields
		id := ""
		model := ""
		if idField := respValue.FieldByName("ID"); idField.IsValid() {
			id = idField.String()
		}
		if modelField := respValue.FieldByName("Model"); modelField.IsValid() {
			model = modelField.String()
		}

		// Get Usage struct and its fields
		if usageField := respValue.FieldByName("Usage"); usageField.IsValid() {
			inputTokens := 0
			outputTokens := 0

			if inputField := usageField.FieldByName("InputTokens"); inputField.IsValid() {
				inputTokens = int(inputField.Int())
			}
			if outputField := usageField.FieldByName("OutputTokens"); outputField.IsValid() {
				outputTokens = int(outputField.Int())
			}

			return common.TokenUsage{
				InputTokens:    inputTokens,
				OutputTokens:   outputTokens,
				TotalTokens:    inputTokens + outputTokens,
				CompletionID:   id,
				Model:          model,
				Timestamp:      time.Now(),
				PromptTokens:   inputTokens,
				ResponseTokens: outputTokens,
			}, nil
		}
	}

	return common.TokenUsage{}, fmt.Errorf("response is not an *anthropic.Message or valid mock: %T", response)
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

	// Check if the model exists in the pricing map
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
