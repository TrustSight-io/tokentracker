package providers

import (
	"testing"

	"github.com/TrustSight-io/tokentracker"
)

func TestGeminiProvider_Name(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewGeminiProvider(config)

	if provider.Name() != "gemini" {
		t.Errorf("GeminiProvider.Name() = %q, expected %q", provider.Name(), "gemini")
	}
}

func TestGeminiProvider_SupportsModel(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewGeminiProvider(config)

	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{
			name:     "Gemini Pro",
			model:    "gemini-pro",
			expected: true,
		},
		{
			name:     "Gemini Ultra",
			model:    "gemini-ultra",
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
				t.Errorf("GeminiProvider.SupportsModel(%q) = %v, expected %v", tt.model, provider.SupportsModel(tt.model), tt.expected)
			}
		})
	}
}

func TestGeminiProvider_CountTokens(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewGeminiProvider(config)

	tests := []struct {
		name         string
		params       tokentracker.TokenCountParams
		wantErr      bool
		minExpected  int
		maxExpected  int
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
				Model: "gemini-pro",
				Text:  StringPtr("This is a simple test text for Gemini tokenization."),
			},
			wantErr:     false,
			minExpected: 5,
			maxExpected: 20,
		},
		{
			name: "Chat messages",
			params: tokentracker.TokenCountParams{
				Model: "gemini-ultra",
				Messages: []tokentracker.Message{
					{
						Role:    "system",
						Content: "You are a helpful assistant.",
					},
					{
						Role:    "user",
						Content: "Tell me about Gemini tokenization.",
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
				Model: "gemini-pro",
			},
			wantErr: true,
		},
		{
			name: "With response tokens estimation",
			params: tokentracker.TokenCountParams{
				Model:              "gemini-ultra",
				Text:               StringPtr("Estimate response tokens"),
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
				t.Errorf("GeminiProvider.CountTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.InputTokens < tt.minExpected || got.InputTokens > tt.maxExpected {
				t.Errorf("GeminiProvider.CountTokens() InputTokens = %v, expected between %v and %v", 
					got.InputTokens, tt.minExpected, tt.maxExpected)
			}

			if tt.params.CountResponseTokens && got.ResponseTokens == 0 {
				t.Errorf("GeminiProvider.CountTokens() ResponseTokens = 0, expected > 0 when CountResponseTokens is true")
			}

			if !tt.params.CountResponseTokens && got.ResponseTokens != 0 {
				t.Errorf("GeminiProvider.CountTokens() ResponseTokens = %v, expected 0 when CountResponseTokens is false", 
					got.ResponseTokens)
			}
		})
	}
}

func TestGeminiProvider_CalculatePrice(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewGeminiProvider(config)

	// Update pricing information to ensure consistent test results
	provider.UpdatePricing()

	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		wantErr      bool
	}{
		{
			name:         "Gemini Pro",
			model:        "gemini-pro",
			inputTokens:  1000,
			outputTokens: 500,
			wantErr:      false,
		},
		{
			name:         "Gemini Ultra",
			model:        "gemini-ultra",
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
				t.Errorf("GeminiProvider.CalculatePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Verify price calculation
			if tt.model == "gemini-pro" {
				expectedInputCost := float64(tt.inputTokens) * 0.00000025
				expectedOutputCost := float64(tt.outputTokens) * 0.0000005
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

func TestGeminiProvider_ExtractTokenUsageFromResponse(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewGeminiProvider(config)

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
				"id":    "gen_123",
				"model": "gemini-pro",
			},
			wantErr: true,
		},
		{
			name: "Response with usage",
			response: map[string]interface{}{
				"id":    "gen_123",
				"model": "gemini-pro",
				"usage": map[string]interface{}{
					"prompt_tokens":     float64(100),
					"completion_tokens": float64(50),
				},
			},
			wantErr:        false,
			expectedInput:  100,
			expectedOutput: 50,
		},
		{
			name: "Response with invalid token counts",
			response: map[string]interface{}{
				"id":    "gen_123",
				"model": "gemini-pro",
				"usage": map[string]interface{}{
					"prompt_tokens":     "invalid",
					"completion_tokens": "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenCount, err := provider.ExtractTokenUsageFromResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeminiProvider.ExtractTokenUsageFromResponse() error = %v, wantErr %v", err, tt.wantErr)
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

func TestGeminiProvider_GetModelInfo(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewGeminiProvider(config)

	tests := []struct {
		name      string
		model     string
		wantErr   bool
		checkInfo bool
	}{
		{
			name:      "Gemini Pro",
			model:     "gemini-pro",
			wantErr:   false,
			checkInfo: true,
		},
		{
			name:      "Gemini Ultra",
			model:     "gemini-ultra",
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
				t.Errorf("GeminiProvider.GetModelInfo() error = %v, wantErr %v", err, tt.wantErr)
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

func TestGeminiProvider_UpdatePricing(t *testing.T) {
	config := tokentracker.NewConfig()
	provider := NewGeminiProvider(config)

	// Clear pricing before test
	config.Providers["gemini"] = tokentracker.ProviderConfig{
		Models: make(map[string]tokentracker.ModelPricing),
	}

	// Update pricing
	err := provider.UpdatePricing()
	if err != nil {
		t.Errorf("GeminiProvider.UpdatePricing() error = %v", err)
		return
	}

	// Check that pricing has been updated
	models := []string{"gemini-pro", "gemini-ultra"}
	for _, model := range models {
		pricing, exists := config.GetModelPricing("gemini", model)
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
