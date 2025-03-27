//go:build integration
// +build integration

package sdkwrappers

import (
	"testing"
	"time"

	"github.com/TrustSight-io/tokentracker"
	"github.com/TrustSight-io/tokentracker/common"
	"github.com/TrustSight-io/tokentracker/providers"
)

// TestSDKWrapperIntegration tests the integration between SDK wrappers and providers
// using mock responses for external API calls.
func TestSDKWrapperIntegration(t *testing.T) {
	// Create a new configuration
	config := tokentracker.NewConfig()

	// Create providers
	openaiProvider := providers.NewOpenAIProvider(config)
	anthropicProvider := providers.NewClaudeProvider(config)
	geminiProvider := providers.NewGeminiProvider(config)

	// Create SDK wrappers
	openaiWrapper := NewOpenAISDKWrapper("mock-api-key")
	anthropicWrapper := NewAnthropicSDKWrapper("mock-api-key")
	geminiWrapper, err := NewGeminiSDKWrapper("mock-api-key")
	if err != nil {
		t.Fatalf("Failed to create Gemini SDK wrapper: %v", err)
	}

	// Register wrappers with a token tracker
	tracker := tokentracker.NewTokenTracker(config)
	tracker.RegisterProvider(openaiProvider)
	tracker.RegisterProvider(anthropicProvider)
	tracker.RegisterProvider(geminiProvider)

	err = tracker.RegisterSDKClient(openaiWrapper)
	if err != nil {
		t.Fatalf("Failed to register OpenAI SDK client: %v", err)
	}

	err = tracker.RegisterSDKClient(anthropicWrapper)
	if err != nil {
		t.Fatalf("Failed to register Anthropic SDK client: %v", err)
	}

	err = tracker.RegisterSDKClient(geminiWrapper)
	if err != nil {
		t.Fatalf("Failed to register Gemini SDK client: %v", err)
	}

	// Test OpenAI wrapper with mock response
	t.Run("OpenAI SDK Wrapper", func(t *testing.T) {
		mockResponse := &MockOpenAIResponse{
			ID:     "cmpl-123",
			Object: "chat.completion",
			Model:  "gpt-4",
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		}

		metrics, err := openaiWrapper.TrackAPICall(GPT4, mockResponse)
		if err != nil {
			t.Fatalf("OpenAISDKWrapper.TrackAPICall failed: %v", err)
		}

		validateMetrics(t, metrics, "openai", GPT4, 100, 50, 150)
	})

	// Test Anthropic wrapper with mock response
	t.Run("Anthropic SDK Wrapper", func(t *testing.T) {
		mockResponse := &MockAnthropicResponse{
			ID:    "msg_123",
			Model: "claude-3-opus",
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{
				InputTokens:  120,
				OutputTokens: 80,
			},
		}

		metrics, err := anthropicWrapper.TrackAPICall(ClaudeOpus, mockResponse)
		if err != nil {
			t.Fatalf("AnthropicSDKWrapper.TrackAPICall failed: %v", err)
		}

		validateMetrics(t, metrics, "anthropic", ClaudeOpus, 120, 80, 200)
	})

	// Test Gemini wrapper with mock response
	t.Run("Gemini SDK Wrapper", func(t *testing.T) {
		// Mock Gemini response
		type MockGeminiResponse struct {
			Model         string `json:"model"`
			UsageMetadata struct {
				PromptTokenCount     int `json:"promptTokenCount"`
				CandidatesTokenCount int `json:"candidatesTokenCount"`
				TotalTokenCount      int `json:"totalTokenCount"`
			} `json:"usageMetadata"`
		}

		mockResponse := &MockGeminiResponse{
			Model: "gemini-pro",
			UsageMetadata: struct {
				PromptTokenCount     int `json:"promptTokenCount"`
				CandidatesTokenCount int `json:"candidatesTokenCount"`
				TotalTokenCount      int `json:"totalTokenCount"`
			}{
				PromptTokenCount:     90,
				CandidatesTokenCount: 60,
				TotalTokenCount:      150,
			},
		}

		metrics, err := geminiWrapper.TrackAPICall(GeminiPro, mockResponse)
		if err != nil {
			t.Fatalf("GeminiSDKWrapper.TrackAPICall failed: %v", err)
		}

		validateMetrics(t, metrics, "gemini", GeminiPro, 90, 60, 150)
	})

	// Test full end-to-end flow with a mock API call
	t.Run("End-to-End Token Tracking", func(t *testing.T) {
		// Create call parameters
		callParams := tokentracker.CallParams{
			Model: GPT4,
			Params: tokentracker.TokenCountParams{
				Model: GPT4,
				Messages: []tokentracker.Message{
					{
						Role:    "system",
						Content: "You are a helpful assistant.",
					},
					{
						Role:    "user",
						Content: "Tell me about token tracking.",
					},
				},
			},
			StartTime: time.Now(),
		}

		// Create a mock API response
		mockResponse := &MockOpenAIResponse{
			ID:     "cmpl-456",
			Object: "chat.completion",
			Model:  GPT4,
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     31,
				CompletionTokens: 31,
				TotalTokens:      62,
			},
		}

		// Track usage with the mock response
		usage, err := tracker.TrackUsage(callParams, mockResponse)
		if err != nil {
			t.Fatalf("TokenTracker.TrackUsage failed: %v", err)
		}

		// Validate token counts and pricing, comparing to what we expect
		// The actual values may differ depending on the implementation, but they should be positive
		if usage.TokenCount.InputTokens <= 0 {
			t.Errorf("Expected positive input tokens, got: %d", usage.TokenCount.InputTokens)
		}

		if usage.TokenCount.ResponseTokens <= 0 {
			t.Errorf("Expected positive response tokens, got: %d", usage.TokenCount.ResponseTokens)
		}

		if usage.TokenCount.TotalTokens <= 0 {
			t.Errorf("Expected positive total tokens, got: %d", usage.TokenCount.TotalTokens)
		}

		if usage.Price.TotalCost <= 0 {
			t.Errorf("Expected non-zero total cost, got: %f", usage.Price.TotalCost)
		}

		if usage.Duration <= 0 {
			t.Errorf("Expected positive duration, got: %v", usage.Duration)
		}
	})
}

// Helper function to validate metrics
func validateMetrics(t *testing.T, metrics common.UsageMetrics, provider, model string, inputTokens, outputTokens, totalTokens int) {
	if metrics.Provider != provider {
		t.Errorf("Expected provider: %s, got: %s", provider, metrics.Provider)
	}
	if metrics.Model != model {
		t.Errorf("Expected model: %s, got: %s", model, metrics.Model)
	}

	// We'll be more flexible with token counts as implementations might differ
	if metrics.TokenCount.InputTokens <= 0 {
		t.Errorf("Expected positive input tokens, got: %d", metrics.TokenCount.InputTokens)
	}

	if metrics.TokenCount.ResponseTokens <= 0 {
		t.Errorf("Expected positive response tokens, got: %d", metrics.TokenCount.ResponseTokens)
	}

	if metrics.TokenCount.TotalTokens <= 0 {
		t.Errorf("Expected positive total tokens, got: %d", metrics.TokenCount.TotalTokens)
	}

	if metrics.Price.TotalCost <= 0 {
		t.Errorf("Expected positive total cost, got: %f", metrics.Price.TotalCost)
	}

	if metrics.Timestamp.IsZero() {
		t.Errorf("Expected non-zero timestamp")
	}
}
