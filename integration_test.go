//go:build integration
// +build integration

package tokentracker_test

import (
	"testing"
	"time"

	"github.com/TrustSight-io/tokentracker"
	"github.com/TrustSight-io/tokentracker/providers"
)

// TestTokenTrackerIntegration tests the full token tracker functionality
// with all providers integrated.
func TestTokenTrackerIntegration(t *testing.T) {
	// Create a new configuration
	config := tokentracker.NewConfig()

	// Create a new token tracker
	tracker := tokentracker.NewTokenTracker(config)

	// Register all providers
	openaiProvider := providers.NewOpenAIProvider(config)
	geminiProvider := providers.NewGeminiProvider(config)
	claudeProvider := providers.NewClaudeProvider(config)

	tracker.RegisterProvider(openaiProvider)
	tracker.RegisterProvider(geminiProvider)
	tracker.RegisterProvider(claudeProvider)

	// Test text token counting across all providers
	t.Run("Text Token Counting", func(t *testing.T) {
		text := "This is a sample text for integration testing of token counting across all providers."
		
		models := []struct {
			name     string
			model    string
			provider string
		}{
			{"OpenAI", "gpt-3.5-turbo", "openai"},
			{"Gemini", "gemini-pro", "gemini"},
			{"Claude", "claude-3-haiku", "anthropic"},
		}

		for _, m := range models {
			t.Run(m.name, func(t *testing.T) {
				params := tokentracker.TokenCountParams{
					Model: m.model,
					Text:  &text,
				}

				tokenCount, err := tracker.CountTokens(params)
				if err != nil {
					t.Fatalf("%s token counting failed: %v", m.name, err)
				}

				if tokenCount.InputTokens <= 0 {
					t.Errorf("%s returned zero or negative input tokens: %d", m.name, tokenCount.InputTokens)
				}
				if tokenCount.Model != m.model {
					t.Errorf("Expected model: %s, got: %s", m.model, tokenCount.Model)
				}
				if tokenCount.Provider != m.provider {
					t.Errorf("Expected provider: %s, got: %s", m.provider, tokenCount.Provider)
				}
			})
		}
	})

	// Test message token counting across all providers
	t.Run("Message Token Counting", func(t *testing.T) {
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
		
		models := []struct {
			name     string
			model    string
			provider string
		}{
			{"OpenAI", "gpt-4", "openai"},
			{"Gemini", "gemini-pro", "gemini"},
			{"Claude", "claude-3-opus", "anthropic"},
		}

		for _, m := range models {
			t.Run(m.name, func(t *testing.T) {
				params := tokentracker.TokenCountParams{
					Model:    m.model,
					Messages: messages,
				}

				tokenCount, err := tracker.CountTokens(params)
				if err != nil {
					t.Fatalf("%s message token counting failed: %v", m.name, err)
				}

				if tokenCount.InputTokens <= 0 {
					t.Errorf("%s returned zero or negative input tokens: %d", m.name, tokenCount.InputTokens)
				}
				if tokenCount.Model != m.model {
					t.Errorf("Expected model: %s, got: %s", m.model, tokenCount.Model)
				}
				if tokenCount.Provider != m.provider {
					t.Errorf("Expected provider: %s, got: %s", m.provider, tokenCount.Provider)
				}
			})
		}
	})

	// Test price calculation across all providers
	t.Run("Price Calculation", func(t *testing.T) {
		models := []struct {
			name         string
			model        string
			inputTokens  int
			outputTokens int
		}{
			{"OpenAI", "gpt-4", 1000, 500},
			{"Gemini", "gemini-pro", 1000, 500},
			{"Claude", "claude-3-opus", 1000, 500},
		}

		for _, m := range models {
			t.Run(m.name, func(t *testing.T) {
				price, err := tracker.CalculatePrice(m.model, m.inputTokens, m.outputTokens)
				if err != nil {
					t.Fatalf("%s price calculation failed: %v", m.name, err)
				}

				if price.InputCost <= 0 {
					t.Errorf("%s returned zero or negative input cost: %f", m.name, price.InputCost)
				}
				if price.OutputCost <= 0 {
					t.Errorf("%s returned zero or negative output cost: %f", m.name, price.OutputCost)
				}
				if price.TotalCost <= 0 {
					t.Errorf("%s returned zero or negative total cost: %f", m.name, price.TotalCost)
				}
				if price.Currency == "" {
					t.Errorf("%s returned empty currency", m.name)
				}
			})
		}
	})

	// Test complete usage tracking
	t.Run("Usage Tracking", func(t *testing.T) {
		// Create test parameters for each provider
		testCases := []struct {
			name     string
			model    string
			provider string
		}{
			{"OpenAI", "gpt-3.5-turbo", "openai"},
			{"Gemini", "gemini-pro", "gemini"},
			{"Claude", "claude-3-haiku", "anthropic"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create call parameters
				callParams := tokentracker.CallParams{
					Model: tc.model,
					Params: tokentracker.TokenCountParams{
						Model: tc.model,
						Text:  stringPtr("This is a test message for usage tracking integration testing."),
					},
					StartTime: time.Now().Add(-1 * time.Second), // Simulate 1 second of processing time
				}

				// Create a mock response with token usage
				var mockResponse interface{}
				switch tc.provider {
				case "openai":
					mockResponse = mockOpenAIResponse(tc.model, 20, 30)
				case "anthropic":
					mockResponse = mockAnthropicResponse(tc.model, 20, 30)
				case "gemini":
					mockResponse = mockGeminiResponse(tc.model, 20, 30)
				}

				// Track usage
				usage, err := tracker.TrackUsage(callParams, mockResponse)
				if err != nil {
					t.Fatalf("%s usage tracking failed: %v", tc.name, err)
				}

				// Validate usage metrics
				validateUsageMetrics(t, usage, tc.model, tc.provider)
			})
		}
	})

	// Test provider discovery and selection
	t.Run("Provider Discovery", func(t *testing.T) {
		models := []struct {
			model            string
			expectedProvider string
		}{
			{"gpt-4", "openai"},
			{"gpt-3.5-turbo", "openai"},
			{"claude-3-haiku", "anthropic"},
			{"claude-3-sonnet", "anthropic"},
			{"claude-3-opus", "anthropic"},
			{"gemini-pro", "gemini"},
			{"gemini-ultra", "gemini"},
		}

		for _, m := range models {
			provider, err := tracker.GetProviderForModel(m.model)
			if err != nil {
				t.Errorf("Failed to get provider for model %s: %v", m.model, err)
				continue
			}

			if provider.Name() != m.expectedProvider {
				t.Errorf("Expected provider %s for model %s, got %s",
					m.expectedProvider, m.model, provider.Name())
			}
		}

		// Test with invalid model
		_, err := tracker.GetProviderForModel("nonexistent-model")
		if err == nil {
			t.Errorf("Expected error for nonexistent model, got nil")
		}
	})
}

// Helper functions for creating mock responses

func stringPtr(s string) *string {
	return &s
}

// Mock OpenAI response with token usage
func mockOpenAIResponse(model string, inputTokens, outputTokens int) interface{} {
	return struct {
		ID     string
		Object string
		Model  string
		Usage  struct {
			PromptTokens     int
			CompletionTokens int
			TotalTokens      int
		}
	}{
		ID:     "cmpl-123",
		Object: "chat.completion",
		Model:  model,
		Usage: struct {
			PromptTokens     int
			CompletionTokens int
			TotalTokens      int
		}{
			PromptTokens:     inputTokens,
			CompletionTokens: outputTokens,
			TotalTokens:      inputTokens + outputTokens,
		},
	}
}

// Mock Anthropic response with token usage
func mockAnthropicResponse(model string, inputTokens, outputTokens int) interface{} {
	return struct {
		ID    string
		Model string
		Usage struct {
			InputTokens  int
			OutputTokens int
		}
	}{
		ID:    "msg_123",
		Model: model,
		Usage: struct {
			InputTokens  int
			OutputTokens int
		}{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
		},
	}
}

// Mock Gemini response with token usage
func mockGeminiResponse(model string, inputTokens, outputTokens int) interface{} {
	return struct {
		Model         string
		UsageMetadata struct {
			PromptTokenCount     int
			CandidatesTokenCount int
			TotalTokenCount      int
		}
	}{
		Model: model,
		UsageMetadata: struct {
			PromptTokenCount     int
			CandidatesTokenCount int
			TotalTokenCount      int
		}{
			PromptTokenCount:     inputTokens,
			CandidatesTokenCount: outputTokens,
			TotalTokenCount:      inputTokens + outputTokens,
		},
	}
}

// Helper function to validate usage metrics
func validateUsageMetrics(t *testing.T, usage tokentracker.UsageMetrics, model, provider string) {
	if usage.Model != model {
		t.Errorf("Expected model: %s, got: %s", model, usage.Model)
	}

	if usage.Provider != provider {
		t.Errorf("Expected provider: %s, got: %s", provider, usage.Provider)
	}

	if usage.TokenCount.InputTokens <= 0 {
		t.Errorf("Expected positive input tokens, got: %d", usage.TokenCount.InputTokens)
	}

	if usage.TokenCount.ResponseTokens <= 0 {
		t.Errorf("Expected positive response tokens, got: %d", usage.TokenCount.ResponseTokens)
	}

	if usage.TokenCount.TotalTokens <= 0 {
		t.Errorf("Expected positive total tokens, got: %d", usage.TokenCount.TotalTokens)
	}

	if usage.Price.InputCost <= 0 {
		t.Errorf("Expected positive input cost, got: %f", usage.Price.InputCost)
	}

	if usage.Price.OutputCost <= 0 {
		t.Errorf("Expected positive output cost, got: %f", usage.Price.OutputCost)
	}

	if usage.Price.TotalCost <= 0 {
		t.Errorf("Expected positive total cost, got: %f", usage.Price.TotalCost)
	}

	if usage.Price.Currency == "" {
		t.Errorf("Expected non-empty currency")
	}

	if usage.Duration <= 0 {
		t.Errorf("Expected positive duration, got: %v", usage.Duration)
	}

	if usage.Timestamp.IsZero() {
		t.Errorf("Expected non-zero timestamp")
	}
}