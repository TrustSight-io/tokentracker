package providers

import (
	"encoding/json"
	"fmt"

	"github.com/TrustSight-io/tokentracker"
	"github.com/pkoukk/tiktoken-go"
)

// OpenAIProvider implements the Provider interface for OpenAI models
type OpenAIProvider struct {
	config *tokentracker.Config
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *tokentracker.Config) *OpenAIProvider {
	return &OpenAIProvider{
		config: config,
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// SupportsModel checks if the provider supports a specific model
func (p *OpenAIProvider) SupportsModel(model string) bool {
	supportedModels := map[string]bool{
		"gpt-3.5-turbo":      true,
		"gpt-3.5-turbo-16k":  true,
		"gpt-4":              true,
		"gpt-4-turbo":        true,
		"gpt-4-32k":          true,
		"gpt-4o":             true,
		"text-embedding-ada": true,
		// Add more models as needed
	}

	return supportedModels[model]
}

// CountTokens counts tokens for the given parameters
func (p *OpenAIProvider) CountTokens(params tokentracker.TokenCountParams) (tokentracker.TokenCount, error) {
	if params.Model == "" {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "model is required", nil)
	}

	// Get the encoding for the model
	encoding, err := p.getEncoding(params.Model)
	if err != nil {
		return tokentracker.TokenCount{}, err
	}

	var inputTokens int

	// Count tokens based on the input type
	if params.Text != nil {
		// Count tokens for text
		inputTokens = len(encoding.Encode(*params.Text, nil, nil))
	} else if len(params.Messages) > 0 {
		// Count tokens for chat messages
		inputTokens, err = p.countMessageTokens(params.Model, params.Messages, params.Tools, params.ToolChoice, encoding)
		if err != nil {
			return tokentracker.TokenCount{}, err
		}
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
func (p *OpenAIProvider) CalculatePrice(model string, inputTokens, outputTokens int) (tokentracker.Price, error) {
	if model == "" {
		return tokentracker.Price{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "model is required", nil)
	}

	// Get pricing information for the model
	pricing, exists := p.config.GetModelPricing("openai", model)
	if !exists {
		return tokentracker.Price{}, tokentracker.NewError(tokentracker.ErrPricingNotFound, fmt.Sprintf("pricing not found for model: %s", model), nil)
	}

	// Calculate costs
	inputCost := float64(inputTokens) * pricing.InputPricePerToken
	outputCost := float64(outputTokens) * pricing.OutputPricePerToken
	totalCost := inputCost + outputCost

	return tokentracker.Price{
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  totalCost,
		Currency:   pricing.Currency,
	}, nil
}

// SetSDKClient sets the provider-specific SDK client
func (p *OpenAIProvider) SetSDKClient(client interface{}) {
	// Store the client for later use
	// In a real implementation, this would be used to make API calls
}

// GetModelInfo returns information about a specific model
func (p *OpenAIProvider) GetModelInfo(model string) (interface{}, error) {
	// In a real implementation, this would return model information
	// For now, we'll just return a simple map
	return map[string]interface{}{
		"name":         model,
		"provider":     "openai",
		"capabilities": []string{"text", "chat", "function-calling"},
	}, nil
}

// ExtractTokenUsageFromResponse extracts token usage from a provider response
func (p *OpenAIProvider) ExtractTokenUsageFromResponse(response interface{}) (tokentracker.TokenCount, error) {
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
func (p *OpenAIProvider) UpdatePricing() error {
	// If we have an SDK client, we could use it to fetch the latest pricing
	// For now, we'll just update with hardcoded values

	// GPT-3.5 Turbo pricing (as of March 2024)
	p.config.SetModelPricing("openai", "gpt-3.5-turbo", tokentracker.ModelPricing{
		InputPricePerToken:  0.0000015,
		OutputPricePerToken: 0.000002,
		Currency:            "USD",
	})

	// GPT-4 pricing (as of March 2024)
	p.config.SetModelPricing("openai", "gpt-4", tokentracker.ModelPricing{
		InputPricePerToken:  0.00003,
		OutputPricePerToken: 0.00006,
		Currency:            "USD",
	})

	// GPT-4 Turbo pricing (as of March 2024)
	p.config.SetModelPricing("openai", "gpt-4-turbo", tokentracker.ModelPricing{
		InputPricePerToken:  0.00001,
		OutputPricePerToken: 0.00003,
		Currency:            "USD",
	})

	return nil
}

// getEncoding returns the encoding for the given model
func (p *OpenAIProvider) getEncoding(model string) (*tiktoken.Tiktoken, error) {
	// Map model to encoding
	encodingName := "cl100k_base" // Default for most newer models

	// Override for specific models if needed
	if model == "text-embedding-ada" {
		encodingName = "r50k_base"
	}

	// Get the encoding
	encoding, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		return nil, tokentracker.NewError(tokentracker.ErrTokenizationFailed, "failed to get encoding", err)
	}

	return encoding, nil
}

// countMessageTokens counts tokens for chat messages
func (p *OpenAIProvider) countMessageTokens(_ string, messages []tokentracker.Message, tools []tokentracker.Tool, toolChoice *tokentracker.ToolChoice, encoding *tiktoken.Tiktoken) (int, error) {
	// Convert messages to JSON for token counting
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		return 0, tokentracker.NewError(tokentracker.ErrTokenizationFailed, "failed to marshal messages", err)
	}

	// Count tokens in the messages JSON
	tokens := len(encoding.Encode(string(messagesJSON), nil, nil))

	// Add tokens for tools if present
	if len(tools) > 0 {
		toolsJSON, err := json.Marshal(tools)
		if err != nil {
			return 0, tokentracker.NewError(tokentracker.ErrTokenizationFailed, "failed to marshal tools", err)
		}

		tokens += len(encoding.Encode(string(toolsJSON), nil, nil))
	}

	// Add tokens for tool choice if present
	if toolChoice != nil {
		toolChoiceJSON, err := json.Marshal(toolChoice)
		if err != nil {
			return 0, tokentracker.NewError(tokentracker.ErrTokenizationFailed, "failed to marshal tool choice", err)
		}

		tokens += len(encoding.Encode(string(toolChoiceJSON), nil, nil))
	}

	// Add tokens for message formatting
	// This is a simplified approach; a real implementation would be more precise
	tokens += 3 // For the message format

	return tokens, nil
}

// estimateResponseTokens estimates the number of response tokens
func (p *OpenAIProvider) estimateResponseTokens(model string, inputTokens int) int {
	// This is a very simplified estimation
	// In a real implementation, this would be more sophisticated
	return tokentracker.EstimateResponseTokens(model, inputTokens)
}
