package providers

import (
	"testing"

	"github.com/TrustSight-io/tokentracker"
)

func TestOpenAIProvider_CountTokens(t *testing.T) {
	// Create a new configuration
	config := tokentracker.NewConfig()

	// Create a new OpenAI provider
	provider := NewOpenAIProvider(config)

	// Test cases
	tests := []struct {
		name          string
		params        tokentracker.TokenCountParams
		wantMinTokens int // Minimum expected tokens
		wantMaxTokens int // Maximum expected tokens
		wantErr       bool
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
				Model: "gpt-3.5-turbo",
				Text:  StringPtr("This is a test."),
			},
			wantMinTokens: 4,  // At least 4 tokens
			wantMaxTokens: 10, // At most 10 tokens
			wantErr:       false,
		},
		{
			name: "Longer text",
			params: tokentracker.TokenCountParams{
				Model: "gpt-4",
				Text:  StringPtr("This is a longer test text that should have more tokens than the previous example."),
			},
			wantMinTokens: 10, // At least 10 tokens
			wantMaxTokens: 25, // At most 25 tokens
			wantErr:       false,
		},
		{
			name: "Simple messages",
			params: tokentracker.TokenCountParams{
				Model: "gpt-3.5-turbo",
				Messages: []tokentracker.Message{
					{
						Role:    "system",
						Content: "You are a helpful assistant.",
					},
					{
						Role:    "user",
						Content: "Hello!",
					},
				},
			},
			wantMinTokens: 10, // At least 10 tokens
			wantMaxTokens: 30, // At most 30 tokens
			wantErr:       false,
		},
		{
			name: "No text or messages",
			params: tokentracker.TokenCountParams{
				Model: "gpt-3.5-turbo",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.CountTokens(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenAIProvider.CountTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.InputTokens < tt.wantMinTokens || got.InputTokens > tt.wantMaxTokens {
				t.Errorf("OpenAIProvider.CountTokens() = %v, want between %v and %v", got.InputTokens, tt.wantMinTokens, tt.wantMaxTokens)
			}
		})
	}
}

func TestOpenAIProvider_CalculatePrice(t *testing.T) {
	// Create a new configuration
	config := tokentracker.NewConfig()

	// Create a new OpenAI provider
	provider := NewOpenAIProvider(config)

	// Test cases
	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		wantErr      bool
	}{
		{
			name:         "Empty model",
			model:        "",
			inputTokens:  100,
			outputTokens: 50,
			wantErr:      true,
		},
		{
			name:         "GPT-3.5-Turbo",
			model:        "gpt-3.5-turbo",
			inputTokens:  1000,
			outputTokens: 500,
			wantErr:      false,
		},
		{
			name:         "GPT-4",
			model:        "gpt-4",
			inputTokens:  1000,
			outputTokens: 500,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.CalculatePrice(tt.model, tt.inputTokens, tt.outputTokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenAIProvider.CalculatePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Check that the price is calculated correctly
			if tt.model == "gpt-3.5-turbo" {
				expectedInputCost := float64(tt.inputTokens) * 0.0000015
				expectedOutputCost := float64(tt.outputTokens) * 0.000002
				expectedTotalCost := expectedInputCost + expectedOutputCost

				if got.InputCost != expectedInputCost {
					t.Errorf("OpenAIProvider.CalculatePrice() InputCost = %v, want %v", got.InputCost, expectedInputCost)
				}
				if got.OutputCost != expectedOutputCost {
					t.Errorf("OpenAIProvider.CalculatePrice() OutputCost = %v, want %v", got.OutputCost, expectedOutputCost)
				}
				if got.TotalCost != expectedTotalCost {
					t.Errorf("OpenAIProvider.CalculatePrice() TotalCost = %v, want %v", got.TotalCost, expectedTotalCost)
				}
			}
		})
	}
}
