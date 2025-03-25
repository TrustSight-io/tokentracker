package providers

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/TrustSight-io/tokentracker"
)

// ClaudeProvider implements the Provider interface for Claude models
type ClaudeProvider struct {
	config *tokentracker.Config
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(config *tokentracker.Config) *ClaudeProvider {
	return &ClaudeProvider{
		config: config,
	}
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
	return tokentracker.EstimateResponseTokens(model, inputTokens)
}

// Note: We can add a helper function here if needed for Claude-specific functionality