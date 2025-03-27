//go:build integration
// +build integration

package providers

import (
	"testing"

	"github.com/TrustSight-io/tokentracker"
)

// TestProviderIntegration_AllProviders tests the integration between all providers
// using mock services for external API calls.
func TestProviderIntegration_AllProviders(t *testing.T) {
	// Create a new configuration
	config := tokentracker.NewConfig()

	// Create all providers
	openaiProvider := NewOpenAIProvider(config)
	geminiProvider := NewGeminiProvider(config)
	claudeProvider := NewClaudeProvider(config)

	// Common test input
	inputText := "This is a sample text for token counting across multiple providers."
	
	// Test all providers with the same text input
	providers := []struct {
		name     string
		provider tokentracker.Provider
		model    string
	}{
		{"OpenAI", openaiProvider, "gpt-3.5-turbo"},
		{"Gemini", geminiProvider, "gemini-pro"},
		{"Claude", claudeProvider, "claude-3-haiku"},
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			// Test token counting
			params := tokentracker.TokenCountParams{
				Model: p.model,
				Text:  StringPtr(inputText),
			}

			tokenCount, err := p.provider.CountTokens(params)
			if err != nil {
				t.Fatalf("%s provider token counting failed: %v", p.name, err)
			}

			// Verify token count is reasonable (not zero)
			if tokenCount.InputTokens <= 0 {
				t.Errorf("%s provider returned zero or negative input tokens: %d", p.name, tokenCount.InputTokens)
			}

			// Test price calculation
			price, err := p.provider.CalculatePrice(p.model, tokenCount.InputTokens, 0)
			if err != nil {
				t.Fatalf("%s provider price calculation failed: %v", p.name, err)
			}

			// Verify price is reasonable (has a value and currency)
			if price.InputCost <= 0 {
				t.Errorf("%s provider returned zero or negative input cost: %f", p.name, price.InputCost)
			}
			if price.Currency == "" {
				t.Errorf("%s provider returned empty currency", p.name)
			}
		})
	}
}

// TestProviderIntegration_CrossProviderComparison tests the consistency of token counting
// between different providers for the same input.
func TestProviderIntegration_CrossProviderComparison(t *testing.T) {
	// Create a new configuration
	config := tokentracker.NewConfig()

	// Create all providers
	openaiProvider := NewOpenAIProvider(config)
	geminiProvider := NewGeminiProvider(config)
	claudeProvider := NewClaudeProvider(config)

	// Test message for all providers
	messages := []tokentracker.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: "Tell me about token counting in language models.",
		},
	}

	// Get token counts from each provider
	openaiParams := tokentracker.TokenCountParams{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
	}
	geminiParams := tokentracker.TokenCountParams{
		Model:    "gemini-pro",
		Messages: messages,
	}
	claudeParams := tokentracker.TokenCountParams{
		Model:    "claude-3-haiku",
		Messages: messages,
	}

	openaiCount, err := openaiProvider.CountTokens(openaiParams)
	if err != nil {
		t.Fatalf("OpenAI token counting failed: %v", err)
	}

	geminiCount, err := geminiProvider.CountTokens(geminiParams)
	if err != nil {
		t.Fatalf("Gemini token counting failed: %v", err)
	}

	claudeCount, err := claudeProvider.CountTokens(claudeParams)
	if err != nil {
		t.Fatalf("Claude token counting failed: %v", err)
	}

	// Verify all providers return a reasonable token count
	if openaiCount.InputTokens <= 0 || geminiCount.InputTokens <= 0 || claudeCount.InputTokens <= 0 {
		t.Errorf("One or more providers returned zero or negative input tokens: OpenAI=%d, Gemini=%d, Claude=%d",
			openaiCount.InputTokens, geminiCount.InputTokens, claudeCount.InputTokens)
	}

	// Verify token counts are within a reasonable range of each other (50% variation allowed)
	// Different tokenizers will give different results, but they should be roughly in the same ballpark
	maxCount := max(openaiCount.InputTokens, max(geminiCount.InputTokens, claudeCount.InputTokens))
	minCount := min(openaiCount.InputTokens, min(geminiCount.InputTokens, claudeCount.InputTokens))

	if float64(minCount) < float64(maxCount)*0.5 {
		t.Errorf("Token count variation too large: OpenAI=%d, Gemini=%d, Claude=%d",
			openaiCount.InputTokens, geminiCount.InputTokens, claudeCount.InputTokens)
	}
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
