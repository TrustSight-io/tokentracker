package sdkwrappers

import (
	"testing"

	"github.com/TrustSight-io/tokentracker"
)

// MockOpenAIProvider is a mock Provider implementation for testing
type MockOpenAIProvider struct {
	name     string
	models   []string
	pricing  map[string]tokentracker.ModelPricing
	client   interface{}
	supports bool
}

func (p *MockOpenAIProvider) Name() string {
	return p.name
}

func (p *MockOpenAIProvider) SupportsModel(model string) bool {
	return p.supports
}

func (p *MockOpenAIProvider) CountTokens(params tokentracker.TokenCountParams) (tokentracker.TokenCount, error) {
	return tokentracker.TokenCount{
		InputTokens:    100,
		ResponseTokens: 50,
		TotalTokens:    150,
	}, nil
}

func (p *MockOpenAIProvider) CalculatePrice(model string, inputTokens, outputTokens int) (tokentracker.Price, error) {
	return tokentracker.Price{
		InputCost:  0.0001,
		OutputCost: 0.0002,
		TotalCost:  0.0003,
		Currency:   "USD",
	}, nil
}

func (p *MockOpenAIProvider) SetSDKClient(client interface{}) {
	p.client = client
}

func (p *MockOpenAIProvider) GetModelInfo(model string) (interface{}, error) {
	return map[string]interface{}{
		"name":     model,
		"provider": p.name,
	}, nil
}

func (p *MockOpenAIProvider) ExtractTokenUsageFromResponse(response interface{}) (tokentracker.TokenCount, error) {
	return tokentracker.TokenCount{
		InputTokens:    100,
		ResponseTokens: 50,
		TotalTokens:    150,
	}, nil
}

func (p *MockOpenAIProvider) UpdatePricing() error {
	return nil
}

// MockOpenAIResponse is a mock response for OpenAI API
type MockOpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func TestOpenAISDKWrapper_GetProviderName(t *testing.T) {
	provider := &MockOpenAIProvider{name: "openai", supports: true}
	// Skip actual client creation in tests
	wrapper := &OpenAISDKWrapper{}

	if wrapper.GetProviderName() != "openai" {
		t.Errorf("OpenAISDKWrapper.GetProviderName() = %q, expected %q", wrapper.GetProviderName(), "openai")
	}
}

func TestOpenAISDKWrapper_GetClient(t *testing.T) {
	provider := &MockOpenAIProvider{name: "openai", supports: true}
	// Skip actual client creation in tests
	wrapper := &OpenAISDKWrapper{}

	client := wrapper.GetClient()
	if client == nil {
		t.Errorf("OpenAISDKWrapper.GetClient() returned nil")
	}
}

func TestOpenAISDKWrapper_GetSupportedModels(t *testing.T) {
	provider := &MockOpenAIProvider{name: "openai", supports: true}
	// Skip actual client creation in tests
	wrapper := &OpenAISDKWrapper{}

	models, err := wrapper.GetSupportedModels()
	if err != nil {
		t.Errorf("OpenAISDKWrapper.GetSupportedModels() error = %v", err)
		return
	}

	if len(models) == 0 {
		t.Errorf("OpenAISDKWrapper.GetSupportedModels() returned empty slice")
	}

	// Check that OpenAI models are included
	expectedModels := []string{"gpt-3.5-turbo", "gpt-4"}
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

func TestOpenAISDKWrapper_ExtractTokenUsageFromResponse(t *testing.T) {
	provider := &MockOpenAIProvider{name: "openai", supports: true}
	// Skip actual client creation in tests
	wrapper := &OpenAISDKWrapper{}

	// Create a mock response
	response := &MockOpenAIResponse{
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

	usage, err := wrapper.ExtractTokenUsageFromResponse(response)
	if err != nil {
		t.Errorf("OpenAISDKWrapper.ExtractTokenUsageFromResponse() error = %v", err)
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
	if usage.Model != "gpt-4" {
		t.Errorf("ExtractTokenUsageFromResponse() Model = %v, want gpt-4", usage.Model)
	}
	if usage.CompletionID != "cmpl-123" {
		t.Errorf("ExtractTokenUsageFromResponse() CompletionID = %v, want cmpl-123", usage.CompletionID)
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

func TestOpenAISDKWrapper_FetchCurrentPricing(t *testing.T) {
	provider := &MockOpenAIProvider{name: "openai", supports: true}
	// Skip actual client creation in tests
	wrapper := &OpenAISDKWrapper{}

	pricing, err := wrapper.FetchCurrentPricing()
	if err != nil {
		t.Errorf("OpenAISDKWrapper.FetchCurrentPricing() error = %v", err)
		return
	}

	if len(pricing) == 0 {
		t.Errorf("FetchCurrentPricing() returned empty pricing map")
	}

	// Check that OpenAI models are included
	expectedModels := []string{"gpt-3.5-turbo", "gpt-4"}
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

func TestOpenAISDKWrapper_TrackAPICall(t *testing.T) {
	provider := &MockOpenAIProvider{name: "openai", supports: true}
	// Skip actual client creation in tests
	wrapper := &OpenAISDKWrapper{}

	// Create a mock response
	response := &MockOpenAIResponse{
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

	// Track API call
	metrics, err := wrapper.TrackAPICall("gpt-4", response)
	if err != nil {
		t.Errorf("OpenAISDKWrapper.TrackAPICall() error = %v", err)
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
	if metrics.Model != "gpt-4" {
		t.Errorf("TrackAPICall() Model = %v, want gpt-4", metrics.Model)
	}
	if metrics.Provider != "openai" {
		t.Errorf("TrackAPICall() Provider = %v, want openai", metrics.Provider)
	}
	if metrics.Timestamp.IsZero() {
		t.Errorf("TrackAPICall() Timestamp is zero")
	}

	// Test with nil response
	_, err = wrapper.TrackAPICall("gpt-4", nil)
	if err == nil {
		t.Errorf("Expected error when tracking API call with nil response")
	}

	// Test with unsupported model
	_, err = wrapper.TrackAPICall("unsupported-model", response)
	if err == nil {
		t.Errorf("Expected error when tracking API call with unsupported model")
	}
}

func TestOpenAISDKWrapper_UpdateProviderPricing(t *testing.T) {
	provider := &MockOpenAIProvider{name: "openai", supports: true}
	// Skip actual client creation in tests
	wrapper := &OpenAISDKWrapper{}

	err := wrapper.UpdateProviderPricing()
	if err != nil {
		t.Errorf("OpenAISDKWrapper.UpdateProviderPricing() error = %v", err)
	}
}

func TestOpenAIConstants(t *testing.T) {
	// Check that constants are defined
	if GPT35Turbo == "" {
		t.Errorf("GPT35Turbo constant is empty")
	}
	if GPT4 == "" {
		t.Errorf("GPT4 constant is empty")
	}

	// Check that constants match expected values
	if GPT35Turbo != "gpt-3.5-turbo" {
		t.Errorf("GPT35Turbo = %q, expected %q", GPT35Turbo, "gpt-3.5-turbo")
	}
	if GPT4 != "gpt-4" {
		t.Errorf("GPT4 = %q, expected %q", GPT4, "gpt-4")
	}
}
