package sdkwrappers

import (
	"testing"

	"github.com/TrustSight-io/tokentracker"
	"github.com/TrustSight-io/tokentracker/common"
)

// MockClaudeProvider is a mock Provider implementation for testing
type MockClaudeProvider struct {
	name     string
	models   []string
	pricing  map[string]tokentracker.ModelPricing
	client   interface{}
	supports bool
}

func (p *MockClaudeProvider) Name() string {
	return p.name
}

func (p *MockClaudeProvider) SupportsModel(model string) bool {
	return p.supports
}

func (p *MockClaudeProvider) CountTokens(params tokentracker.TokenCountParams) (tokentracker.TokenCount, error) {
	return tokentracker.TokenCount{
		InputTokens:    100,
		ResponseTokens: 50,
		TotalTokens:    150,
	}, nil
}

func (p *MockClaudeProvider) CalculatePrice(model string, inputTokens, outputTokens int) (tokentracker.Price, error) {
	return tokentracker.Price{
		InputCost:  0.0001,
		OutputCost: 0.0002,
		TotalCost:  0.0003,
		Currency:   "USD",
	}, nil
}

func (p *MockClaudeProvider) SetSDKClient(client interface{}) {
	p.client = client
}

func (p *MockClaudeProvider) GetModelInfo(model string) (interface{}, error) {
	return map[string]interface{}{
		"name":     model,
		"provider": p.name,
	}, nil
}

func (p *MockClaudeProvider) ExtractTokenUsageFromResponse(response interface{}) (tokentracker.TokenCount, error) {
	return tokentracker.TokenCount{
		InputTokens:    100,
		ResponseTokens: 50,
		TotalTokens:    150,
	}, nil
}

func (p *MockClaudeProvider) UpdatePricing() error {
	return nil
}

// MockAnthropicResponse is a mock response for Anthropic API
type MockAnthropicResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func TestAnthropicSDKWrapper_GetProviderName(t *testing.T) {
	// The providers are no longer directly passed to the constructor
	wrapper := &AnthropicSDKWrapper{
		// Add a mock client to avoid nil pointer errors
		client: struct{}{},
	}

	if wrapper.GetProviderName() != "anthropic" {
		t.Errorf("AnthropicSDKWrapper.GetProviderName() = %q, expected %q", wrapper.GetProviderName(), "anthropic")
	}
}

func TestAnthropicSDKWrapper_GetClient(t *testing.T) {
	// The providers are no longer directly passed to the constructor
	wrapper := &AnthropicSDKWrapper{
		// Add a mock client to avoid nil pointer errors
		client: struct{}{},
	}

	client := wrapper.GetClient()
	if client == nil {
		t.Errorf("AnthropicSDKWrapper.GetClient() returned nil")
	}
}

func TestAnthropicSDKWrapper_GetSupportedModels(t *testing.T) {
	// The providers are no longer directly passed to the constructor
	wrapper := &AnthropicSDKWrapper{
		// Add a mock client to avoid nil pointer errors
		client: struct{}{},
	}

	models, err := wrapper.GetSupportedModels()
	if err != nil {
		t.Errorf("AnthropicSDKWrapper.GetSupportedModels() error = %v", err)
		return
	}

	if len(models) == 0 {
		t.Errorf("AnthropicSDKWrapper.GetSupportedModels() returned empty slice")
	}

	// Check that Claude models are included
	expectedModels := []string{"claude-3-haiku", "claude-3-sonnet", "claude-3-opus"}
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

func TestAnthropicSDKWrapper_ExtractTokenUsageFromResponse(t *testing.T) {
	// The providers are no longer directly passed to the constructor
	wrapper := &AnthropicSDKWrapper{}

	// Create a mock response
	response := &MockAnthropicResponse{
		ID:    "msg_123",
		Model: "claude-3-opus",
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{
			InputTokens:  100,
			OutputTokens: 50,
		},
	}

	usage, err := wrapper.ExtractTokenUsageFromResponse(response)
	if err != nil {
		t.Errorf("AnthropicSDKWrapper.ExtractTokenUsageFromResponse() error = %v", err)
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
	if usage.Model != "claude-3-opus" {
		t.Errorf("ExtractTokenUsageFromResponse() Model = %v, want claude-3-opus", usage.Model)
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

func TestAnthropicSDKWrapper_FetchCurrentPricing(t *testing.T) {
	// The providers are no longer directly passed to the constructor
	wrapper := &AnthropicSDKWrapper{}

	pricing, err := wrapper.FetchCurrentPricing()
	if err != nil {
		t.Errorf("AnthropicSDKWrapper.FetchCurrentPricing() error = %v", err)
		return
	}

	if len(pricing) == 0 {
		t.Errorf("FetchCurrentPricing() returned empty pricing map")
	}

	// Check that Claude models are included
	expectedModels := []string{"claude-3-haiku", "claude-3-sonnet", "claude-3-opus"}
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

func TestAnthropicSDKWrapper_TrackAPICall(t *testing.T) {
	// The providers are no longer directly passed to the constructor
	wrapper := &AnthropicSDKWrapper{}

	// Create a mock response
	response := &MockAnthropicResponse{
		ID:    "msg_123",
		Model: "claude-3-opus",
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{
			InputTokens:  100,
			OutputTokens: 50,
		},
	}

	// Track API call
	metrics, err := wrapper.TrackAPICall("claude-3-opus", response)
	if err != nil {
		t.Errorf("AnthropicSDKWrapper.TrackAPICall() error = %v", err)
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
	if metrics.Model != "claude-3-opus" {
		t.Errorf("TrackAPICall() Model = %v, want claude-3-opus", metrics.Model)
	}
	if metrics.Provider != "anthropic" {
		t.Errorf("TrackAPICall() Provider = %v, want anthropic", metrics.Provider)
	}
	if metrics.Timestamp.IsZero() {
		t.Errorf("TrackAPICall() Timestamp is zero")
	}

	// Test with nil response
	_, err = wrapper.TrackAPICall("claude-3-opus", nil)
	if err == nil {
		t.Errorf("Expected error when tracking API call with nil response")
	}

	// Test with unsupported model
	_, err = wrapper.TrackAPICall("unsupported-model", response)
	if err == nil {
		t.Errorf("Expected error when tracking API call with unsupported model")
	}
}

func TestAnthropicSDKWrapper_UpdateProviderPricing(t *testing.T) {
	// The providers are no longer directly passed to the constructor
	wrapper := &AnthropicSDKWrapper{}

	err := wrapper.UpdateProviderPricing()
	if err != nil {
		t.Errorf("AnthropicSDKWrapper.UpdateProviderPricing() error = %v", err)
	}
}

func TestAnthropicConstants(t *testing.T) {
	// Check that constants are defined
	if ClaudeHaiku == "" {
		t.Errorf("ClaudeHaiku constant is empty")
	}
	if ClaudeSonnet == "" {
		t.Errorf("ClaudeSonnet constant is empty")
	}
	if ClaudeOpus == "" {
		t.Errorf("ClaudeOpus constant is empty")
	}

	// Check that constants match expected values
	if ClaudeHaiku != "claude-3-haiku" {
		t.Errorf("ClaudeHaiku = %q, expected %q", ClaudeHaiku, "claude-3-haiku")
	}
	if ClaudeSonnet != "claude-3-sonnet" {
		t.Errorf("ClaudeSonnet = %q, expected %q", ClaudeSonnet, "claude-3-sonnet")
	}
	if ClaudeOpus != "claude-3-opus" {
		t.Errorf("ClaudeOpus = %q, expected %q", ClaudeOpus, "claude-3-opus")
	}
}
