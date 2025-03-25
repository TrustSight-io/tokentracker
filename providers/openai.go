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
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrTokenizationFailed, "failed to get encoding", err)
	}

	var inputTokens int

	// Count tokens based on input type
	if params.Text != nil {
		// Count tokens for text
		inputTokens = len(encoding.Encode(*params.Text, nil, nil))
	} else if len(params.Messages) > 0 {
		// Count tokens for messages
		inputTokens, err = p.countMessageTokens(encoding, params.Messages, params.Tools, params.ToolChoice)
		if err != nil {
			return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrTokenizationFailed, "failed to count message tokens", err)
		}
	} else {
		return tokentracker.TokenCount{}, tokentracker.NewError(tokentracker.ErrInvalidParams, "either text or messages must be provided", nil)
	}

	// Estimate response tokens if requested
	var responseTokens int
	if params.CountResponseTokens {
		// This is a simplified estimation
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
	pricing, exists := p.config.GetModelPricing("openai", model)
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

// getEncoding returns the tiktoken encoding for the given model
func (p *OpenAIProvider) getEncoding(model string) (*tiktoken.Tiktoken, error) {
	var encodingName string

	// Determine the encoding name based on the model
	switch {
	case model == "text-embedding-ada":
		encodingName = "cl100k_base"
	case model == "gpt-3.5-turbo" || model == "gpt-3.5-turbo-16k":
		encodingName = "cl100k_base"
	case model == "gpt-4" || model == "gpt-4-32k" || model == "gpt-4-turbo" || model == "gpt-4o":
		encodingName = "cl100k_base"
	default:
		// Default to cl100k_base for unknown models
		encodingName = "cl100k_base"
	}

	encoding, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		return nil, err
	}

	return encoding, nil
}

// countMessageTokens counts tokens for chat messages
// Implementation based on OpenAI's token counting logic
func (p *OpenAIProvider) countMessageTokens(encoding *tiktoken.Tiktoken, messages []tokentracker.Message, tools []tokentracker.Tool, toolChoice *tokentracker.ToolChoice) (int, error) {
	// Base tokens for the messages format
	tokens := 3 // Every reply is primed with <|start|>assistant<|message|>

	// Count tokens for each message
	for _, message := range messages {
		// Add tokens for message role
		tokens += 4 // Every message follows <|start|>{role}<|message|>

		// Count tokens for content
		switch content := message.Content.(type) {
		case string:
			tokens += len(encoding.Encode(content, nil, nil))
		case []tokentracker.ContentPart:
			for _, part := range content {
				if part.Type == "text" {
					tokens += len(encoding.Encode(part.Text, nil, nil))
				} else if part.Type == "image" {
					// Simplified image token counting
					// This is a placeholder and should be replaced with actual image token counting logic
					tokens += 1000 // Placeholder value
				}
			}
		case []interface{}:
			// Handle array of content parts from JSON
			for _, partInterface := range content {
				if part, ok := partInterface.(map[string]interface{}); ok {
					if partType, ok := part["type"].(string); ok {
						if partType == "text" {
							if text, ok := part["text"].(string); ok {
								tokens += len(encoding.Encode(text, nil, nil))
							}
						} else if partType == "image" {
							// Simplified image token counting
							tokens += 1000 // Placeholder value
						}
					}
				}
			}
		default:
			// Try to handle as JSON
			contentBytes, err := json.Marshal(content)
			if err == nil {
				tokens += len(encoding.Encode(string(contentBytes), nil, nil))
			}
		}
	}

	// Count tokens for tools if provided
	if len(tools) > 0 {
		// Convert tools to JSON for token counting
		toolsJSON, err := json.Marshal(tools)
		if err == nil {
			tokens += len(encoding.Encode(string(toolsJSON), nil, nil))
		}

		// Add base tokens for tools
		tokens += 10 // Placeholder value
	}

	// Count tokens for tool choice if provided
	if toolChoice != nil {
		// Convert tool choice to JSON for token counting
		toolChoiceJSON, err := json.Marshal(toolChoice)
		if err == nil {
			tokens += len(encoding.Encode(string(toolChoiceJSON), nil, nil))
		}
	}

	return tokens, nil
}

// estimateResponseTokens estimates the number of response tokens
func (p *OpenAIProvider) estimateResponseTokens(model string, inputTokens int) int {
	// This is a very simplified estimation
	// In a real implementation, this would be more sophisticated
	return tokentracker.EstimateResponseTokens(model, inputTokens)
}