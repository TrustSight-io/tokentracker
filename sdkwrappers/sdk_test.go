package sdkwrappers

import (
	"testing"
	"time"

	"github.com/TrustSight-io/tokentracker"
)

// MockResponse is a simple mock response for testing
type MockResponse struct {
	ID    string
	Model string
	Usage struct {
		InputTokens  int64
		OutputTokens int64
	}
}

// MockSDKWrapper is a mock implementation of SDKClientWrapper for testing
type MockSDKWrapper struct {
	providerName string
	pricing      map[string]tokentracker.ModelPricing
}

func (m *MockSDKWrapper) GetProviderName() string {
	return m.providerName
}

func (m *MockSDKWrapper) GetClient() interface{} {
	return nil
}

func (m *MockSDKWrapper) GetSupportedModels() ([]string, error) {
	return []string{"model-1", "model-2"}, nil
}

func (m *MockSDKWrapper) ExtractTokenUsageFromResponse(response interface{}) (TokenUsage, error) {
	mockResp, ok := response.(*MockResponse)
	if !ok {
		return TokenUsage{}, nil
	}
	
	return TokenUsage{
		InputTokens:    int(mockResp.Usage.InputTokens),
		OutputTokens:   int(mockResp.Usage.OutputTokens),
		TotalTokens:    int(mockResp.Usage.InputTokens + mockResp.Usage.OutputTokens),
		CompletionID:   mockResp.ID,
		Model:          mockResp.Model,
		Timestamp:      time.Now(),
		PromptTokens:   int(mockResp.Usage.InputTokens),
		ResponseTokens: int(mockResp.Usage.OutputTokens),
	}, nil
}

func (m *MockSDKWrapper) FetchCurrentPricing() (map[string]tokentracker.ModelPricing, error) {
	return m.pricing, nil
}

func (m *MockSDKWrapper) UpdateProviderPricing() error {
	return nil
}

func (m *MockSDKWrapper) TrackAPICall(model string, response interface{}) (tokentracker.UsageMetrics, error) {
	return tokentracker.UsageMetrics{}, nil
}

func TestSDKWrapperInterface(t *testing.T) {
	// Table-driven tests for different wrappers
	tests := []struct {
		name           string
		wrapper        SDKClientWrapper
		expectedName   string
		mockResponse   *MockResponse
		expectedInput  int
		expectedOutput int
	}{
		{
			name: "OpenAI Wrapper",
			wrapper: &MockSDKWrapper{
				providerName: "openai",
				pricing: map[string]tokentracker.ModelPricing{
					"gpt-4": {
						InputPricePerToken:  0.00003,
						OutputPricePerToken: 0.00006,
						Currency:            "USD",
					},
				},
			},
			expectedName: "openai",
			mockResponse: &MockResponse{
				ID:    "cmpl-123",
				Model: "gpt-4",
				Usage: struct {
					InputTokens  int64
					OutputTokens int64
				}{
					InputTokens:  100,
					OutputTokens: 50,
				},
			},
			expectedInput:  100,
			expectedOutput: 50,
		},
		{
			name: "Anthropic Wrapper",
			wrapper: &MockSDKWrapper{
				providerName: "anthropic",
				pricing: map[string]tokentracker.ModelPricing{
					"claude-3-opus": {
						InputPricePerToken:  0.00001,
						OutputPricePerToken: 0.00003,
						Currency:            "USD",
					},
				},
			},
			expectedName: "anthropic",
			mockResponse: &MockResponse{
				ID:    "msg_123",
				Model: "claude-3-opus",
				Usage: struct {
					InputTokens  int64
					OutputTokens int64
				}{
					InputTokens:  200,
					OutputTokens: 75,
				},
			},
			expectedInput:  200,
			expectedOutput: 75,
		},
		{
			name: "Gemini Wrapper",
			wrapper: &MockSDKWrapper{
				providerName: "gemini",
				pricing: map[string]tokentracker.ModelPricing{
					"gemini-pro": {
						InputPricePerToken:  0.00000025,
						OutputPricePerToken: 0.0000005,
						Currency:            "USD",
					},
				},
			},
			expectedName: "gemini",
			mockResponse: &MockResponse{
				ID:    "gen_123",
				Model: "gemini-pro",
				Usage: struct {
					InputTokens  int64
					OutputTokens int64
				}{
					InputTokens:  150,
					OutputTokens: 60,
				},
			},
			expectedInput:  150,
			expectedOutput: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test GetProviderName
			if name := tt.wrapper.GetProviderName(); name != tt.expectedName {
				t.Errorf("GetProviderName() = %v, want %v", name, tt.expectedName)
			}

			// Test FetchCurrentPricing
			pricing, err := tt.wrapper.FetchCurrentPricing()
			if err != nil {
				t.Errorf("FetchCurrentPricing() error = %v", err)
			}
			if len(pricing) == 0 {
				t.Error("FetchCurrentPricing() returned empty pricing map")
			}

			// Test ExtractTokenUsageFromResponse
			usage, err := tt.wrapper.ExtractTokenUsageFromResponse(tt.mockResponse)
			if err != nil {
				t.Errorf("ExtractTokenUsageFromResponse() error = %v", err)
			}
			if usage.InputTokens != tt.expectedInput {
				t.Errorf("ExtractTokenUsageFromResponse() input tokens = %v, want %v", 
					usage.InputTokens, tt.expectedInput)
			}
			if usage.OutputTokens != tt.expectedOutput {
				t.Errorf("ExtractTokenUsageFromResponse() output tokens = %v, want %v", 
					usage.OutputTokens, tt.expectedOutput)
			}
			if usage.TotalTokens != tt.expectedInput+tt.expectedOutput {
				t.Errorf("ExtractTokenUsageFromResponse() total tokens = %v, want %v", 
					usage.TotalTokens, tt.expectedInput+tt.expectedOutput)
			}
			if usage.Model != tt.mockResponse.Model {
				t.Errorf("ExtractTokenUsageFromResponse() model = %v, want %v", 
					usage.Model, tt.mockResponse.Model)
			}
			if usage.CompletionID != tt.mockResponse.ID {
				t.Errorf("ExtractTokenUsageFromResponse() completion ID = %v, want %v", 
					usage.CompletionID, tt.mockResponse.ID)
			}
		})
	}
}

// TestAnthropicSDKWrapper tests the actual AnthropicSDKWrapper implementation
func TestAnthropicSDKWrapper(t *testing.T) {
	// Skip this test if we don't have a real API key
	t.Skip("Skipping test that requires a real API key")

	// Create a mock provider
	provider := &MockProvider{
		name: "anthropic",
	}

	// Create a wrapper with a fake API key
	wrapper := NewAnthropicSDKWrapper("fake-api-key", provider)

	// Test GetProviderName
	if name := wrapper.GetProviderName(); name != "anthropic" {
		t.Errorf("GetProviderName() = %v, want %v", name, "anthropic")
	}

	// Test GetSupportedModels
	models, err := wrapper.GetSupportedModels()
	if err != nil {
		t.Errorf("GetSupportedModels() error = %v", err)
	}
	if len(models) == 0 {
		t.Error("GetSupportedModels() returned empty models list")
	}

	// Test FetchCurrentPricing
	pricing, err := wrapper.FetchCurrentPricing()
	if err != nil {
		t.Errorf("FetchCurrentPricing() error = %v", err)
	}
	if len(pricing) == 0 {
		t.Error("FetchCurrentPricing() returned empty pricing map")
	}
}

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	name string
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) CountTokens(params tokentracker.TokenCountParams) (tokentracker.TokenCount, error) {
	return tokentracker.TokenCount{}, nil
}

func (p *MockProvider) CalculatePrice(model string, inputTokens, outputTokens int) (tokentracker.Price, error) {
	return tokentracker.Price{}, nil
}

func (p *MockProvider) SupportsModel(model string) bool {
	return true
}

func (p *MockProvider) SetSDKClient(client interface{}) {
	// Do nothing
}

func (p *MockProvider) GetModelInfo(model string) (interface{}, error) {
	return nil, nil
}

func (p *MockProvider) ExtractTokenUsageFromResponse(response interface{}) (tokentracker.TokenCount, error) {
	return tokentracker.TokenCount{}, nil
}

func (p *MockProvider) UpdatePricing() error {
	return nil
}
