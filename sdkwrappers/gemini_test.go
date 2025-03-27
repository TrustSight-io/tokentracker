package sdkwrappers

import (
	"testing"

	"github.com/TrustSight-io/tokentracker"
)

// MockGeminiProvider is a mock Provider implementation for testing
type MockGeminiProvider struct {
	name     string
	models   []string
	pricing  map[string]tokentracker.ModelPricing
	client   interface{}
	supports bool
}

func (p *MockGeminiProvider) Name() string {
	return p.name
}

func (p *MockGeminiProvider) SupportsModel(model string) bool {
	return p.supports
}

func (p *MockGeminiProvider) CountTokens(params tokentracker.TokenCountParams) (tokentracker.TokenCount, error) {
	return tokentracker.TokenCount{
		InputTokens:    100,
		ResponseTokens: 50,
		TotalTokens:    150,
	}, nil
}

func (p *MockGeminiProvider) CalculatePrice(model string, inputTokens, outputTokens int) (tokentracker.Price, error) {
	return tokentracker.Price{
		InputCost:  0.0001,
		OutputCost: 0.0002,
		TotalCost:  0.0003,
		Currency:   "USD",
	}, nil
}

func (p *MockGeminiProvider) SetSDKClient(client interface{}) {
	p.client = client
}

func (p *MockGeminiProvider) GetModelInfo(model string) (interface{}, error) {
	return map[string]interface{}{
		"name":     model,
		"provider": p.name,
	}, nil
}

func (p *MockGeminiProvider) ExtractTokenUsageFromResponse(response interface{}) (tokentracker.TokenCount, error) {
	return tokentracker.TokenCount{
		InputTokens:    100,
		ResponseTokens: 50,
		TotalTokens:    150,
	}, nil
}

func (p *MockGeminiProvider) UpdatePricing() error {
	return nil
}

// MockGeminiResponse is a mock response for Gemini API
type MockGeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func TestGeminiSDKWrapper_GetProviderName(t *testing.T) {
	provider := &MockGeminiProvider{name: "gemini", supports: true}
	wrapper := NewGeminiSDKWrapper("test-api-key", provider)

	if wrapper.GetProviderName() != "gemini" {
		t.Errorf("GeminiSDKWrapper.GetProviderName() = %q, expected %q", wrapper.GetProviderName(), "gemini")
	}
}

func TestGeminiSDKWrapper_GetClient(t *testing.T) {
	provider := &MockGeminiProvider{name: "gemini", supports: true}
	wrapper := NewGeminiSDKWrapper("test-api-key", provider)

	client := wrapper.GetClient()
	if client == nil {
		t.Errorf("GeminiSDKWrapper.GetClient() returned nil")
	}
}

func TestGeminiSDKWrapper_GetSupportedModels(t *testing.T) {
	provider := &MockGeminiProvider{name: "gemini", supports: true}
	wrapper := NewGeminiSDKWrapper("test-api-key", provider)

	models, err := wrapper.GetSupportedModels()
	if err != nil {
		t.Errorf("GeminiSDKWrapper.GetSupportedModels() error = %v", err)
		return
	}

	if len(models) == 0 {
		t.Errorf("GeminiSDKWrapper.GetSupportedModels() returned empty slice")
	}

	// Check that Gemini models are included
	expectedModels := []string{"gemini-pro", "gemini-ultra"}
	foundModels := make(map[string]bool)
	for _, model := range models {
		foundModels[model] = true
	}

	for _, expectedModel := range expectedModels {
		if !foundModels[expectedModel] {
			t.Errorf("Expected model %q not found in supported models", expectedModel)
		}
	}
}

func TestGeminiSDKWrapper_ExtractTokenUsageFromResponse(t *testing.T) {
	provider := &MockGeminiProvider{name: "gemini", supports: true}
	wrapper := NewGeminiSDKWrapper("test-api-key", provider)

	// Create a mock response
	response := &MockGeminiResponse{
		UsageMetadata: struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		}{
			PromptTokenCount:     100,
			CandidatesTokenCount: 50,
			TotalTokenCount:      150,
		},
	}

	usage, err := wrapper.ExtractTokenUsageFromResponse(response)
	if err != nil {
		t.Errorf("GeminiSDKWrapper.ExtractTokenUsageFromResponse() error = %v", err)
		return
	}

	// Check extracted usage
	if usage.InputTokens != 100 {
		t.Errorf("ExtractTokenUsageFromResponse() InputTokens = %v, want 100", usage.InputTokens)
	}
	if usage.OutputTokens != 50 {
		t.Errorf("ExtractTokenUsageFromResponse() OutputTokens = %v, want 50", usage.OutputTokens)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("ExtractTokenUsageFromResponse() TotalTokens = %v, want 150", usage.TotalTokens)
	}

	// Test with nil response
	_, err = wrapper.ExtractTokenUsageFromResponse(nil)
	if err == nil {
		t.Errorf("Expected error when extracting token usage from nil response")
	}

	// Test with unsupported response type
	_, err = wrapper.ExtractTokenUsageFromResponse("string response")
	if err == nil {
		t.Errorf("Expected error when extracting token usage from unsupported response type")
	}
}

func TestGeminiSDKWrapper_FetchCurrentPricing(t *testing.T) {
	provider := &MockGeminiProvider{name: "gemini", supports: true}
	wrapper := NewGeminiSDKWrapper("test-api-key", provider)

	pricing, err := wrapper.FetchCurrentPricing()
	if err != nil {
		t.Errorf("GeminiSDKWrapper.FetchCurrentPricing() error = %v", err)
		return
	}

	if len(pricing) == 0 {
		t.Errorf("FetchCurrentPricing() returned empty pricing map")
	}

	// Check that Gemini models are included
	expectedModels := []string{"gemini-pro", "gemini-ultra"}
	for _, model := range expectedModels {
		if _, exists := pricing[model]; !exists {
			t.Errorf("Expected pricing for model %q not found", model)
		}
	}

	// Check pricing values
	for model, modelPricing := range pricing {
		if modelPricing.InputPricePerToken <= 0 {
			t.Errorf("FetchCurrentPricing() InputPricePerToken for %s = %v, expected > 0", model, modelPricing.InputPricePerToken)
		}
		if modelPricing.OutputPricePerToken <= 0 {
			t.Errorf("FetchCurrentPricing() OutputPricePerToken for %s = %v, expected > 0", model, modelPricing.OutputPricePerToken)
		}
		if modelPricing.Currency != "USD" {
			t.Errorf("FetchCurrentPricing() Currency for %s = %v, expected USD", model, modelPricing.Currency)
		}
	}
}

func TestGeminiSDKWrapper_TrackAPICall(t *testing.T) {
	provider := &MockGeminiProvider{name: "gemini", supports: true}
	wrapper := NewGeminiSDKWrapper("test-api-key", provider)

	// Create a mock response
	response := &MockGeminiResponse{
		UsageMetadata: struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		}{
			PromptTokenCount:     100,
			CandidatesTokenCount: 50,
			TotalTokenCount:      150,
		},
	}

	// Track API call
	metrics, err := wrapper.TrackAPICall("gemini-pro", response)
	if err != nil {
		t.Errorf("GeminiSDKWrapper.TrackAPICall() error = %v", err)
		return
	}

	// Check metrics
	if metrics.TokenCount.InputTokens != 100 {
		t.Errorf("TrackAPICall() InputTokens = %v, want 100", metrics.TokenCount.InputTokens)
	}
	if metrics.TokenCount.ResponseTokens != 50 {
		t.Errorf("TrackAPICall() ResponseTokens = %v, want 50", metrics.TokenCount.ResponseTokens)
	}
	if metrics.TokenCount.TotalTokens != 150 {
		t.Errorf("TrackAPICall() TotalTokens = %v, want 150", metrics.TokenCount.TotalTokens)
	}
	if metrics.Price.TotalCost <= 0 {
		t.Errorf("TrackAPICall() TotalCost = %v, expected > 0", metrics.Price.TotalCost)
	}
	if metrics.Model != "gemini-pro" {
		t.Errorf("TrackAPICall() Model = %v, want gemini-pro", metrics.Model)
	}
	if metrics.Provider != "gemini" {
		t.Errorf("TrackAPICall() Provider = %v, want gemini", metrics.Provider)
	}
	if metrics.Timestamp.IsZero() {
		t.Errorf("TrackAPICall() Timestamp is zero")
	}

	// Test with nil response
	_, err = wrapper.TrackAPICall("gemini-pro", nil)
	if err == nil {
		t.Errorf("Expected error when tracking API call with nil response")
	}

	// Test with unsupported model
	_, err = wrapper.TrackAPICall("unsupported-model", response)
	if err == nil {
		t.Errorf("Expected error when tracking API call with unsupported model")
	}
}

func TestGeminiSDKWrapper_UpdateProviderPricing(t *testing.T) {
	provider := &MockGeminiProvider{name: "gemini", supports: true}
	wrapper := NewGeminiSDKWrapper("test-api-key", provider)

	err := wrapper.UpdateProviderPricing()
	if err != nil {
		t.Errorf("GeminiSDKWrapper.UpdateProviderPricing() error = %v", err)
	}
}

func TestGeminiConstants(t *testing.T) {
	// Check that constants are defined
	if GeminiPro == "" {
		t.Errorf("GeminiPro constant is empty")
	}
	if GeminiUltra == "" {
		t.Errorf("GeminiUltra constant is empty")
	}

	// Check that constants match expected values
	if GeminiPro != "gemini-pro" {
		t.Errorf("GeminiPro = %q, expected %q", GeminiPro, "gemini-pro")
	}
	if GeminiUltra != "gemini-ultra" {
		t.Errorf("GeminiUltra = %q, expected %q", GeminiUltra, "gemini-ultra")
	}
}
