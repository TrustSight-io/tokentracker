package sdkwrappers_test

import (
	"testing"
	"time"

	"github.com/TrustSight-io/tokentracker/common"
	"github.com/TrustSight-io/tokentracker/sdkwrappers"
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
	pricing      map[string]common.ModelPricing
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

func (m *MockSDKWrapper) ExtractTokenUsageFromResponse(response interface{}) (common.TokenUsage, error) {
	mockResp, ok := response.(*MockResponse)
	if !ok {
		return common.TokenUsage{}, nil
	}
	
	return common.TokenUsage{
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

func (m *MockSDKWrapper) FetchCurrentPricing() (map[string]common.ModelPricing, error) {
	return m.pricing, nil
}

// Additional methods required by the SDKClientWrapper interface
func (m *MockSDKWrapper) UpdateProviderPricing() error {
	return nil
}

func (m *MockSDKWrapper) TrackAPICall(model string, response interface{}) (common.UsageMetrics, error) {
	return common.UsageMetrics{}, nil
}

func TestSDKWrapperInterface(t *testing.T) {
	// Table-driven tests for different wrappers
	tests := []struct {
		name           string
		wrapper        sdkwrappers.SDKClientWrapper
		expectedName   string
		mockResponse   *MockResponse
		expectedInput  int
		expectedOutput int
	}{
		{
			name: "OpenAI Wrapper",
			wrapper: &MockSDKWrapper{
				providerName: "openai",
				pricing: map[string]common.ModelPricing{
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
				pricing: map[string]common.ModelPricing{
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
				pricing: map[string]common.ModelPricing{
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
