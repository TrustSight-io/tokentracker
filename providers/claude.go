package providers

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/TrustSight-io/tokentracker"
)

// ClaudeProvider implements the Provider interface for Claude models
type ClaudeProvider struct {
	config    *tokentracker.Config
	sdkClient interface{}
	modelInfo map[string]interface{}
	mu        sync.RWMutex
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(config *tokentracker.Config) *ClaudeProvider {
	provider := &ClaudeProvider{
		config:    config,
		modelInfo: make(map[string]interface{}),
	}

	// Initialize with default model info
	provider.initializeModelInfo()

	return provider
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return "anthropic"
}

// SupportsModel checks if the provider supports a specific model
func (p *ClaudeProvider) SupportsModel(model string) bool {
	supportedModels := map[string]bool{
		"claude-3-haiku":  true,
		"claude-3-sonnet": true,
		"claude-3-opus":   true,
		// Add more models as needed
	}

	return supportedModels[model]
}

// CountTokens counts tokens for the given parameters
// Note: This is a simplified implementation for Claude token counting
// In a production environment, you would want to use Anthropic's official tokenizer
func (p *ClaudeProvider) CountTokens(params tokentracker.TokenCountParams) (tokentracker.TokenCount, error) {
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
func (p *ClaudeProvider) CalculatePrice(model string, inputTokens, outputTokens int) (tokentracker.Price, error) {
	pricing, exists := p.config.GetModelPricing("anthropic", model)
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

// SetSDKClient sets the provider-specific SDK client
func (p *ClaudeProvider) SetSDKClient(client interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sdkClient = client
}

// GetModelInfo returns information about a specific model
func (p *ClaudeProvider) GetModelInfo(model string) (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	info, exists := p.modelInfo[model]
	if !exists {
		return nil, tokentracker.NewError(tokentracker.ErrInvalidModel, fmt.Sprintf("model info not found for: %s", model), nil)
	}

	return info, nil
}

// ExtractTokenUsageFromResponse extracts token usage from a provider response
func (p *ClaudeProvider) ExtractTokenUsageFromResponse(response interface{}) (tokentracker.TokenCount, error) {
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
	// The exact structure depends on the Anthropic API response format
	usage, ok := respMap["usage"].(map[string]interface{})
	if !ok {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "usage information not found in response", nil)
	}

	// Extract token counts
	inputTokens, ok1 := usage["input_tokens"].(float64)
	outputTokens, ok2 := usage["output_tokens"].(float64)

	if !ok1 || !ok2 {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "token counts not found in response", nil)
	}

	return tokentracker.TokenCount{
		InputTokens:    int(inputTokens),
		ResponseTokens: int(outputTokens),
		TotalTokens:    int(inputTokens) + int(outputTokens),
	}, nil
}

// UpdatePricing updates the pricing information for this provider
func (p *ClaudeProvider) UpdatePricing() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If we have an SDK client, we could use it to fetch the latest pricing
	// For now, we'll just update with hardcoded values
	
	// Claude 3 Haiku pricing (as of March 2024)
	p.config.SetModelPricing("anthropic", "claude-3-haiku", tokentracker.ModelPricing{
		InputPricePerToken:  0.00000025,
		OutputPricePerToken: 0.00000125,
		Currency:            "USD",
	})

	// Claude 3 Sonnet pricing
	p.config.SetModelPricing("anthropic", "claude-3-sonnet", tokentracker.ModelPricing{
		InputPricePerToken:  0.000003,
		OutputPricePerToken: 0.000015,
		Currency:            "USD",
	})

	// Claude 3 Opus pricing
	p.config.SetModelPricing("anthropic", "claude-3-opus", tokentracker.ModelPricing{
		InputPricePerToken:  0.000015,
		OutputPricePerToken: 0.000075,
		Currency:            "USD",
	})

	return nil
}

// approximateTokenCount provides an approximate token count for Claude models
// This is a simplified implementation and should be replaced with Anthropic's official tokenizer
func (p *ClaudeProvider) approximateTokenCount(text string) int {
	// Check if we have a cached result
	if count, exists := tokentracker.GetCachedTokenCount("anthropic", "", text); exists {
		return count
	}

	// Claude uses a tokenizer similar to GPT models but with some differences
	// A rough approximation is about 4 characters per token for English text
	// This is a very simplified approach and should be replaced with a proper tokenizer
	charCount := utf8.RuneCountInString(text)
	
	// Claude tends to have slightly fewer tokens than GPT for the same text
	tokenCount := (charCount * 95) / 400 // Approximately 0.95 * charCount / 4

	// Add a small overhead for special tokens
	tokenCount += 5

	// Cache the result
	tokentracker.SetCachedTokenCount("anthropic", "", text, tokenCount)

	return tokenCount
}

// countMessageTokens counts tokens for chat messages
func (p *ClaudeProvider) countMessageTokens(messages []tokentracker.Message, tools []tokentracker.Tool, toolChoice *tokentracker.ToolChoice) int {
	// Extract all text from messages
	allText := tokentracker.ExtractTextFromMessages(messages)

	// Count tokens for the combined text
	tokens := p.approximateTokenCount(allText)

	// Add tokens for message structure (roles, formatting)
	// Claude has specific formatting for messages
	tokens += len(messages) * 6 // Claude has slightly more overhead per message

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
func (p *ClaudeProvider) estimateResponseTokens(model string, inputTokens int) int {
	// This is a simplified estimation based on the model
	switch {
	case strings.Contains(model, "opus"):
		return inputTokens * 2 // Claude Opus can be quite verbose
	case strings.Contains(model, "sonnet"):
		return inputTokens * 3 / 2 // Claude Sonnet is moderately verbose
	case strings.Contains(model, "haiku"):
		return inputTokens // Claude Haiku is more concise
	default:
		return tokentracker.EstimateResponseTokens(model, inputTokens)
	}
}

// initializeModelInfo initializes the model information
func (p *ClaudeProvider) initializeModelInfo() {
	p.modelInfo["claude-3-haiku"] = map[string]interface{}{
		"contextWindow": 200000,
		"description":   "Claude 3 Haiku - fastest and most compact model",
	}

	p.modelInfo["claude-3-sonnet"] = map[string]interface{}{
		"contextWindow": 200000,
		"description":   "Claude 3 Sonnet - balanced performance and intelligence",
	}

	p.modelInfo["claude-3-opus"] = map[string]interface{}{
		"contextWindow": 200000,
		"description":   "Claude 3 Opus - most powerful model for complex tasks",
	}
}
