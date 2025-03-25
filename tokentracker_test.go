package tokentracker

import (
	"testing"
	"time"
)

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	name           string
	supportedModel string
	tokenCount     TokenCount
	price          Price
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) SupportsModel(model string) bool {
	return model == p.supportedModel
}

func (p *MockProvider) CountTokens(params TokenCountParams) (TokenCount, error) {
	if params.Model != p.supportedModel {
		return TokenCount{}, NewError(ErrInvalidModel, "unsupported model", nil)
	}
	return p.tokenCount, nil
}

func (p *MockProvider) CalculatePrice(model string, inputTokens, outputTokens int) (Price, error) {
	if model != p.supportedModel {
		return Price{}, NewError(ErrInvalidModel, "unsupported model", nil)
	}
	return p.price, nil
}

func TestDefaultTokenTracker_CountTokens(t *testing.T) {
	// Create a new configuration
	config := NewConfig()

	// Create a new token tracker
	tracker := NewTokenTracker(config)

	// Register a mock provider
	mockProvider := &MockProvider{
		name:           "mock",
		supportedModel: "mock-model",
		tokenCount: TokenCount{
			InputTokens:    100,
			ResponseTokens: 50,
			TotalTokens:    150,
		},
		price: Price{
			InputCost:  0.0001,
			OutputCost: 0.0002,
			TotalCost:  0.0003,
			Currency:   "USD",
		},
	}
	tracker.RegisterProvider(mockProvider)

	// Test cases
	tests := []struct {
		name    string
		params  TokenCountParams
		want    TokenCount
		wantErr bool
	}{
		{
			name: "Empty model",
			params: TokenCountParams{
				Text: stringPtr("Test text"),
			},
			wantErr: true,
		},
		{
			name: "Unsupported model",
			params: TokenCountParams{
				Model: "unsupported-model",
				Text:  stringPtr("Test text"),
			},
			wantErr: true,
		},
		{
			name: "Supported model",
			params: TokenCountParams{
				Model: "mock-model",
				Text:  stringPtr("Test text"),
			},
			want: TokenCount{
				InputTokens:    100,
				ResponseTokens: 50,
				TotalTokens:    150,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tracker.CountTokens(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultTokenTracker.CountTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.InputTokens != tt.want.InputTokens {
				t.Errorf("DefaultTokenTracker.CountTokens() InputTokens = %v, want %v", got.InputTokens, tt.want.InputTokens)
			}
			if got.ResponseTokens != tt.want.ResponseTokens {
				t.Errorf("DefaultTokenTracker.CountTokens() ResponseTokens = %v, want %v", got.ResponseTokens, tt.want.ResponseTokens)
			}
			if got.TotalTokens != tt.want.TotalTokens {
				t.Errorf("DefaultTokenTracker.CountTokens() TotalTokens = %v, want %v", got.TotalTokens, tt.want.TotalTokens)
			}
		})
	}
}

func TestDefaultTokenTracker_CalculatePrice(t *testing.T) {
	// Create a new configuration
	config := NewConfig()

	// Create a new token tracker
	tracker := NewTokenTracker(config)

	// Register a mock provider
	mockProvider := &MockProvider{
		name:           "mock",
		supportedModel: "mock-model",
		tokenCount: TokenCount{
			InputTokens:    100,
			ResponseTokens: 50,
			TotalTokens:    150,
		},
		price: Price{
			InputCost:  0.0001,
			OutputCost: 0.0002,
			TotalCost:  0.0003,
			Currency:   "USD",
		},
	}
	tracker.RegisterProvider(mockProvider)

	// Test cases
	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		want         Price
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
			name:         "Unsupported model",
			model:        "unsupported-model",
			inputTokens:  100,
			outputTokens: 50,
			wantErr:      true,
		},
		{
			name:         "Supported model",
			model:        "mock-model",
			inputTokens:  100,
			outputTokens: 50,
			want: Price{
				InputCost:  0.0001,
				OutputCost: 0.0002,
				TotalCost:  0.0003,
				Currency:   "USD",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tracker.CalculatePrice(tt.model, tt.inputTokens, tt.outputTokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultTokenTracker.CalculatePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.InputCost != tt.want.InputCost {
				t.Errorf("DefaultTokenTracker.CalculatePrice() InputCost = %v, want %v", got.InputCost, tt.want.InputCost)
			}
			if got.OutputCost != tt.want.OutputCost {
				t.Errorf("DefaultTokenTracker.CalculatePrice() OutputCost = %v, want %v", got.OutputCost, tt.want.OutputCost)
			}
			if got.TotalCost != tt.want.TotalCost {
				t.Errorf("DefaultTokenTracker.CalculatePrice() TotalCost = %v, want %v", got.TotalCost, tt.want.TotalCost)
			}
			if got.Currency != tt.want.Currency {
				t.Errorf("DefaultTokenTracker.CalculatePrice() Currency = %v, want %v", got.Currency, tt.want.Currency)
			}
		})
	}
}

func TestDefaultTokenTracker_TrackUsage(t *testing.T) {
	// Create a new configuration
	config := NewConfig()

	// Create a new token tracker
	tracker := NewTokenTracker(config)

	// Register a mock provider
	mockProvider := &MockProvider{
		name:           "mock",
		supportedModel: "mock-model",
		tokenCount: TokenCount{
			InputTokens:    100,
			ResponseTokens: 50,
			TotalTokens:    150,
		},
		price: Price{
			InputCost:  0.0001,
			OutputCost: 0.0002,
			TotalCost:  0.0003,
			Currency:   "USD",
		},
	}
	tracker.RegisterProvider(mockProvider)

	// Test cases
	tests := []struct {
		name       string
		callParams CallParams
		response   interface{}
		wantErr    bool
	}{
		{
			name: "Empty model",
			callParams: CallParams{
				Model: "",
				Params: TokenCountParams{
					Text: stringPtr("Test text"),
				},
				StartTime: time.Now(),
			},
			response: nil,
			wantErr:  true,
		},
		{
			name: "Unsupported model",
			callParams: CallParams{
				Model: "unsupported-model",
				Params: TokenCountParams{
					Model: "unsupported-model",
					Text:  stringPtr("Test text"),
				},
				StartTime: time.Now(),
			},
			response: nil,
			wantErr:  true,
		},
		{
			name: "Supported model",
			callParams: CallParams{
				Model: "mock-model",
				Params: TokenCountParams{
					Model: "mock-model",
					Text:  stringPtr("Test text"),
				},
				StartTime: time.Now().Add(-time.Second),
			},
			response: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tracker.TrackUsage(tt.callParams, tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultTokenTracker.TrackUsage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Check that the usage metrics are populated correctly
			if got.TokenCount.InputTokens != mockProvider.tokenCount.InputTokens {
				t.Errorf("DefaultTokenTracker.TrackUsage() InputTokens = %v, want %v", got.TokenCount.InputTokens, mockProvider.tokenCount.InputTokens)
			}
			if got.Price.TotalCost != mockProvider.price.TotalCost {
				t.Errorf("DefaultTokenTracker.TrackUsage() TotalCost = %v, want %v", got.Price.TotalCost, mockProvider.price.TotalCost)
			}
			if got.Provider != mockProvider.name {
				t.Errorf("DefaultTokenTracker.TrackUsage() Provider = %v, want %v", got.Provider, mockProvider.name)
			}
			if got.Duration < time.Second {
				t.Errorf("DefaultTokenTracker.TrackUsage() Duration = %v, want at least 1s", got.Duration)
			}
		})
	}
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}