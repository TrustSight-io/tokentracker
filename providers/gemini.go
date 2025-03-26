package providers

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/TrustSight-io/tokentracker"
)

// GeminiProvider implements the Provider interface for Gemini models
type GeminiProvider struct {
	config *tokentracker.Config
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(config *tokentracker.Config) *GeminiProvider {
	return &GeminiProvider{
		config: config,
	}
}

// Name returns the provider name
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// SupportsModel checks if the provider supports a specific model
func (p *GeminiProvider) SupportsModel(model string) bool {
	supportedModels := map[string]bool{
		"gemini-pro":   true,
		"gemini-ultra": true,
		// Add more models as needed
	}

	return supportedModels[model]
}

// CountTokens counts tokens for the given parameters
// Note: This is a simplified implementation for Gemini token counting
// In a production environment, you would want to use Google's official tokenizer
func (p *GeminiProvider) CountTokens(params tokentracker.TokenCountParams) (tokentracker.TokenCount, error) {
	if params.Model == "" {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "model is required", nil)
	}

	var inputTokens int

	// Count tokens based on input type
	if params.Text != nil {
		// Count tokens for text
		inputTokens = p.approximateTokenCount(*params.Text)
	} else if len(params.Messages) > 0 {
		// Count tokens for messages
		inputTokens = p.countMessageTokens(params.Messages, params.Tools, params.ToolChoice)
	} else {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "either text or messages must be provided", nil)
	}

	// Estimate response tokens if requested
	var responseTokens int
	if params.CountResponseTokens {
		responseTokens = p.estimateResponseTokens(params.Model, inputTokens)
	}

	return tokentracker.TokenCount{
		InputTokens:    inputTokens,
		ResponseTokens: responseTokens,
		TotalTokens:    inputTokens + responseTokens,
	}, nil
}

// CalculatePrice calculates price based on token usage
func (p *GeminiProvider) CalculatePrice(model string, inputTokens, outputTokens int) (tokentracker.Price, error) {
	pricing, exists := p.config.GetModelPricing("gemini", model)
	if !exists {
		return tokentracker.Price{}, tokentracker.NewError(tokentracker.ErrPricingNotFound, fmt.Sprintf("pricing not found for model: %s", model), nil)
	}

	inputCost := float64(inputTokens) * pricing.InputPricePerToken
	outputCost := float64(outputTokens) * pricing.OutputPricePerToken

	return tokentracker.Price{
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  inputCost + outputCost,
		Currency:   pricing.Currency,
	}, nil
}

// approximateTokenCount provides an approximate token count for Gemini models
// This is a simplified implementation and should be replaced with Google's official tokenizer
func (p *GeminiProvider) approximateTokenCount(text string) int {
	// Check if we have a cached result
	if count, exists := tokentracker.GetCachedTokenCount("gemini", "", text); exists {
		return count
	}

	// Gemini uses a tokenizer similar to GPT models
	// A rough approximation is about 4 characters per token for English text
	// This is a very simplified approach and should be replaced with a proper tokenizer
	charCount := utf8.RuneCountInString(text)
	tokenCount := charCount / 4

	// Add a small overhead for special tokens
	tokenCount += 3

	// Cache the result
	tokentracker.SetCachedTokenCount("gemini", "", text, tokenCount)

	return tokenCount
}

// countMessageTokens counts tokens for chat messages
func (p *GeminiProvider) countMessageTokens(messages []tokentracker.Message, tools []tokentracker.Tool, toolChoice *tokentracker.ToolChoice) int {
	// Extract all text from messages
	allText := tokentracker.ExtractTextFromMessages(messages)

	// Count tokens for the combined text
	tokens := p.approximateTokenCount(allText)

	// Add tokens for message structure (roles, formatting)
	tokens += len(messages) * 4

	// Count tokens for tools if provided
	if len(tools) > 0 {
		toolsJSON, err := json.Marshal(tools)
		if err == nil {
			tokens += p.approximateTokenCount(string(toolsJSON))
		}
	}

	// Count tokens for tool choice if provided
	if toolChoice != nil {
		toolChoiceJSON, err := json.Marshal(toolChoice)
		if err == nil {
			tokens += p.approximateTokenCount(string(toolChoiceJSON))
		}
	}

	return tokens
}

// estimateResponseTokens estimates the number of response tokens
func (p *GeminiProvider) estimateResponseTokens(model string, inputTokens int) int {
	return tokentracker.EstimateResponseTokens(model, inputTokens)
}

// SetSDKClient sets the provider-specific SDK client
func (p *GeminiProvider) SetSDKClient(client interface{}) {
	// Store the client for later use
	// In a real implementation, this would be used to make API calls
}

// GetModelInfo returns information about a specific model
func (p *GeminiProvider) GetModelInfo(model string) (interface{}, error) {
	// In a real implementation, this would return model information
	// For now, we'll just return a simple map
	return map[string]interface{}{
		"name":         model,
		"provider":     "gemini",
		"capabilities": []string{"text", "chat", "image-understanding"},
	}, nil
}

// ExtractTokenUsageFromResponse extracts token usage from a provider response
func (p *GeminiProvider) ExtractTokenUsageFromResponse(response interface{}) (tokentracker.TokenCount, error) {
	// Check if response is nil
	if response == nil {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "response is nil", nil)
	}

	// Try to cast to map[string]interface{} which is common for JSON responses
	respMap, ok := response.(map[string]interface{})
	if !ok {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "response is not a map", nil)
	}

	// Extract usage information from the response
	usage, ok := respMap["usage"].(map[string]interface{})
	if !ok {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "usage information not found in response", nil)
	}

	// Extract token counts
	promptTokens, ok1 := usage["prompt_tokens"].(float64)
	completionTokens, ok2 := usage["completion_tokens"].(float64)
	totalTokens, ok3 := usage["total_tokens"].(float64)

	if !ok1 || !ok2 || !ok3 {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "token counts not found in response", nil)
	}

	return tokentracker.TokenCount{
		InputTokens:    int(promptTokens),
		ResponseTokens: int(completionTokens),
		TotalTokens:    int(totalTokens),
	}, nil
}

// UpdatePricing updates the pricing information for this provider
func (p *GeminiProvider) UpdatePricing() error {
	// If we have an SDK client, we could use it to fetch the latest pricing
	// For now, we'll just update with hardcoded values
	
	// Gemini Pro pricing (as of March 2024)
	p.config.SetModelPricing("gemini", "gemini-pro", tokentracker.ModelPricing{
		InputPricePerToken:  0.00000025,
		OutputPricePerToken: 0.0000005,
		Currency:            "USD",
	})
	
	// Gemini Ultra pricing (as of March 2024)
	p.config.SetModelPricing("gemini", "gemini-ultra", tokentracker.ModelPricing{
		InputPricePerToken:  0.00001,
		OutputPricePerToken: 0.00003,
		Currency:            "USD",
	})
	
	return nil
}