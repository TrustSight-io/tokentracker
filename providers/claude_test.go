package providers

import (
	"testing"

	"github.com/TrustSight-io/tokentracker"
)

func TestClaudeProvider_Name(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewClaudeProvider(config)

	if provider.Name() != "anthropic" {
		t.Errorf("ClaudeProvider.Name() = %q, expected %q", provider.Name(), "anthropic")
	}
}

func TestClaudeProvider_SupportsModel(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewClaudeProvider(config)

	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{
			name:     "Claude Haiku",
			model:    "claude-3-haiku",
			expected: true,
		},
		{
			name:     "Claude Sonnet",
			model:    "claude-3-sonnet",
			expected: true,
		},
		{
			name:     "Claude Opus",
			model:    "claude-3-opus",
			expected: true,
		},
		{
			name:     "Unsupported model",
			model:    "gpt-4",
			expected: false,
		},
		{
			name:     "Empty model",
			model:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if provider.SupportsModel(tt.model) != tt.expected {
				t.Errorf("ClaudeProvider.SupportsModel(%q) = %v, expected %v", tt.model, provider.SupportsModel(tt.model), tt.expected)
			}
		})
	}
}

func TestClaudeProvider_CountTokens(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewClaudeProvider(config)

	tests := []struct {
		name        string
		params      tokentracker.TokenCountParams
		wantErr     bool
		minExpected int
		maxExpected int
	}{
		{
			name: "Empty model",
			params: tokentracker.TokenCountParams{
				Text: StringPtr("Test text"),
			},
			wantErr: true,
		},
		{
			name: "Simple text",
			params: tokentracker.TokenCountParams{
				Model: "claude-3-haiku",
				Text:  StringPtr("This is a simple test text for Claude tokenization."),
			},
			wantErr:     false,
			minExpected: 5,
			maxExpected: 20,
		},
		{
			name: "Chat messages",
			params: tokentracker.TokenCountParams{
				Model: "claude-3-opus",
				Messages: []tokentracker.Message{
					{
						Role:    "system",
						Content: "You are a helpful assistant.",
					},
					{
						Role:    "user",
						Content: "Tell me about Claude tokenization.",
					},
				},
			},
			wantErr:     false,
			minExpected: 10,
			maxExpected: 30,
		},
		{
			name: "No text or messages",
			params: tokentracker.TokenCountParams{
				Model: "claude-3-sonnet",
			},
			wantErr: true,
		},
		{
			name: "With tools",
			params: tokentracker.TokenCountParams{
				Model: "claude-3-opus",
				Text:  StringPtr("Test with tools"),
				Tools: []tokentracker.Tool{
					{
						Type: "function",
						Function: map[string]string{
							"name":        "get_weather",
							"description": "Get the weather for a location",
						},
					},
				},
			},
			wantErr:     false,
			minExpected: 5,
			maxExpected: 50,
		},
		{
			name: "With response tokens estimation",
			params: tokentracker.TokenCountParams{
				Model:               "claude-3-sonnet",
				Text:                StringPtr("Estimate response tokens"),
				CountResponseTokens: true,
			},
			wantErr:     false,
			minExpected: 3,
			maxExpected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.CountTokens(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaudeProvider.CountTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.InputTokens < tt.minExpected || got.InputTokens > tt.maxExpected {
				t.Errorf("ClaudeProvider.CountTokens() InputTokens = %v, expected between %v and %v",
					got.InputTokens, tt.minExpected, tt.maxExpected)
			}

			if tt.params.CountResponseTokens && got.ResponseTokens == 0 {
				t.Errorf("ClaudeProvider.CountTokens() ResponseTokens = 0, expected > 0 when CountResponseTokens is true")
			}

			if !tt.params.CountResponseTokens && got.ResponseTokens != 0 {
				t.Errorf("ClaudeProvider.CountTokens() ResponseTokens = %v, expected 0 when CountResponseTokens is false",
					got.ResponseTokens)
			}
		})
	}
}

func TestClaudeProvider_CalculatePrice(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewClaudeProvider(config)

	// Update pricing information to ensure consistent test results
	err := provider.UpdatePricing()
	if err != nil {
		t.Fatalf("Failed to update pricing: %v", err)
	}

	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		wantErr      bool
	}{
		{
			name:         "Claude Haiku",
			model:        "claude-3-haiku",
			inputTokens:  1000,
			outputTokens: 500,
			wantErr:      false,
		},
		{
			name:         "Claude Sonnet",
			model:        "claude-3-sonnet",
			inputTokens:  1000,
			outputTokens: 500,
			wantErr:      false,
		},
		{
			name:         "Claude Opus",
			model:        "claude-3-opus",
			inputTokens:  1000,
			outputTokens: 500,
			wantErr:      false,
		},
		{
			name:         "Unsupported model",
			model:        "unsupported-model",
			inputTokens:  1000,
			outputTokens: 500,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := provider.CalculatePrice(tt.model, tt.inputTokens, tt.outputTokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaudeProvider.CalculatePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Verify price calculation
			if tt.model == "claude-3-haiku" {
				expectedInputCost := float64(tt.inputTokens) * 0.00000025
				expectedOutputCost := float64(tt.outputTokens) * 0.00000125
				expectedTotalCost := expectedInputCost + expectedOutputCost

				if price.InputCost != expectedInputCost {
					t.Errorf("CalculatePrice() InputCost = %v, want %v", price.InputCost, expectedInputCost)
				}
				if price.OutputCost != expectedOutputCost {
					t.Errorf("CalculatePrice() OutputCost = %v, want %v", price.OutputCost, expectedOutputCost)
				}
				if price.TotalCost != expectedTotalCost {
					t.Errorf("CalculatePrice() TotalCost = %v, want %v", price.TotalCost, expectedTotalCost)
				}
				if price.Currency != "USD" {
					t.Errorf("CalculatePrice() Currency = %v, want USD", price.Currency)
				}
			}
		})
	}
}

func TestClaudeProvider_ExtractTokenUsageFromResponse(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewClaudeProvider(config)

	tests := []struct {
		name           string
		response       interface{}
		wantErr        bool
		expectedInput  int
		expectedOutput int
	}{
		{
			name:     "Nil response",
			response: nil,
			wantErr:  true,
		},
		{
			name:     "Non-map response",
			response: "string response",
			wantErr:  true,
		},
		{
			name: "Response without usage",
			response: map[string]interface{}{
				"id":    "msg_123",
				"model": "claude-3-opus",
			},
			wantErr: true,
		},
		{
			name: "Response with usage",
			response: map[string]interface{}{
				"id":    "msg_123",
				"model": "claude-3-opus",
				"usage": map[string]interface{}{
					"input_tokens":  float64(100),
					"output_tokens": float64(50),
				},
			},
			wantErr:        false,
			expectedInput:  100,
			expectedOutput: 50,
		},
		{
			name: "Response with invalid token counts",
			response: map[string]interface{}{
				"id":    "msg_123",
				"model": "claude-3-opus",
				"usage": map[string]interface{}{
					"input_tokens":  "invalid",
					"output_tokens": "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenCount, err := provider.ExtractTokenUsageFromResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaudeProvider.ExtractTokenUsageFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if tokenCount.InputTokens != tt.expectedInput {
				t.Errorf("ExtractTokenUsageFromResponse() InputTokens = %v, want %v", tokenCount.InputTokens, tt.expectedInput)
			}
			if tokenCount.ResponseTokens != tt.expectedOutput {
				t.Errorf("ExtractTokenUsageFromResponse() ResponseTokens = %v, want %v", tokenCount.ResponseTokens, tt.expectedOutput)
			}
			if tokenCount.TotalTokens != tt.expectedInput+tt.expectedOutput {
				t.Errorf("ExtractTokenUsageFromResponse() TotalTokens = %v, want %v", tokenCount.TotalTokens, tt.expectedInput+tt.expectedOutput)
			}
		})
	}
}

func TestClaudeProvider_GetModelInfo(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewClaudeProvider(config)

	tests := []struct {
		name      string
		model     string
		wantErr   bool
		checkInfo bool
	}{
		{
			name:      "Claude Haiku",
			model:     "claude-3-haiku",
			wantErr:   false,
			checkInfo: true,
		},
		{
			name:      "Claude Sonnet",
			model:     "claude-3-sonnet",
			wantErr:   false,
			checkInfo: true,
		},
		{
			name:      "Claude Opus",
			model:     "claude-3-opus",
			wantErr:   false,
			checkInfo: true,
		},
		{
			name:    "Unsupported model",
			model:   "unsupported-model",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := provider.GetModelInfo(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaudeProvider.GetModelInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if tt.checkInfo {
				modelInfo, ok := info.(map[string]interface{})
				if !ok {
					t.Errorf("GetModelInfo() did not return a map[string]interface{}")
					return
				}

				// Check context window
				contextWindow, exists := modelInfo["contextWindow"]
				if !exists {
					t.Errorf("GetModelInfo() did not include contextWindow")
				} else if contextWindow.(int) <= 0 {
					t.Errorf("GetModelInfo() contextWindow = %v, expected > 0", contextWindow)
				}

				// Check description
				description, exists := modelInfo["description"]
				if !exists {
					t.Errorf("GetModelInfo() did not include description")
				} else if description.(string) == "" {
					t.Errorf("GetModelInfo() description is empty")
				}
			}
		})
	}
}

func TestClaudeProvider_UpdatePricing(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewClaudeProvider(config)

	// Clear pricing before test
	config.Providers["anthropic"] = tokentracker.ProviderConfig{
		Models: make(map[string]tokentracker.ModelPricing),
	}

	// Update pricing
	err := provider.UpdatePricing()
	if err != nil {
		t.Errorf("ClaudeProvider.UpdatePricing() error = %v", err)
		return
	}

	// Check that pricing has been updated
	models := []string{"claude-3-haiku", "claude-3-sonnet", "claude-3-opus"}
	for _, model := range models {
		pricing, exists := config.GetModelPricing("anthropic", model)
		if !exists {
			t.Errorf("UpdatePricing() did not set pricing for model %s", model)
			continue
		}

		if pricing.InputPricePerToken <= 0 {
			t.Errorf("UpdatePricing() InputPricePerToken = %v, expected > 0", pricing.InputPricePerToken)
		}
		if pricing.OutputPricePerToken <= 0 {
			t.Errorf("UpdatePricing() OutputPricePerToken = %v, expected > 0", pricing.OutputPricePerToken)
		}
		if pricing.Currency != "USD" {
			t.Errorf("UpdatePricing() Currency = %v, expected USD", pricing.Currency)
		}
	}
}
