package tokentracker

import (
	"fmt"
	"sync"
	"testing"
)

// MockSimpleProvider is a minimal implementation of the Provider interface for testing
type MockSimpleProvider struct {
	name            string
	supportedModels map[string]bool
}

func (p *MockSimpleProvider) Name() string {
	return p.name
}

func (p *MockSimpleProvider) SupportsModel(model string) bool {
	return p.supportedModels[model]
}

func (p *MockSimpleProvider) CountTokens(params TokenCountParams) (TokenCount, error) {
	return TokenCount{}, nil
}

func (p *MockSimpleProvider) CalculatePrice(model string, inputTokens, outputTokens int) (Price, error) {
	return Price{}, nil
}

func (p *MockSimpleProvider) SetSDKClient(client interface{}) {
	// No-op for mock
}

func (p *MockSimpleProvider) GetModelInfo(model string) (interface{}, error) {
	return nil, nil
}

func (p *MockSimpleProvider) ExtractTokenUsageFromResponse(response interface{}) (TokenCount, error) {
	return TokenCount{}, nil
}

func (p *MockSimpleProvider) UpdatePricing() error {
	return nil
}

func TestNewProviderRegistry(t *testing.T) {
	registry := NewProviderRegistry()
	if registry == nil {
		t.Fatal("NewProviderRegistry() returned nil")
	}
	if registry.providers == nil {
		t.Error("Expected providers map to be initialized")
	}
	if len(registry.providers) != 0 {
		t.Errorf("Expected providers map to be empty, got %d providers", len(registry.providers))
	}
}

func TestProviderRegistry_Register(t *testing.T) {
	registry := NewProviderRegistry()
	provider := &MockSimpleProvider{
		name: "test-provider",
		supportedModels: map[string]bool{
			"model-1": true,
			"model-2": true,
		},
	}

	// Register the provider
	registry.Register(provider)

	// Check that the provider was registered correctly
	if len(registry.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(registry.providers))
	}

	// Try to get the provider
	p, exists := registry.Get("test-provider")
	if !exists {
		t.Error("Expected provider to exist in registry")
	}
	if p != provider {
		t.Errorf("Expected Get() to return registered provider")
	}
}

func TestProviderRegistry_Get(t *testing.T) {
	registry := NewProviderRegistry()
	provider1 := &MockSimpleProvider{name: "provider-1", supportedModels: map[string]bool{"model-1": true}}
	provider2 := &MockSimpleProvider{name: "provider-2", supportedModels: map[string]bool{"model-2": true}}

	registry.Register(provider1)
	registry.Register(provider2)

	tests := []struct {
		name          string
		providerName  string
		expectedFound bool
		expectedProvider Provider
	}{
		{
			name:          "Existing provider",
			providerName:  "provider-1",
			expectedFound: true,
			expectedProvider: provider1,
		},
		{
			name:          "Another existing provider",
			providerName:  "provider-2",
			expectedFound: true,
			expectedProvider: provider2,
		},
		{
			name:          "Non-existent provider",
			providerName:  "provider-3",
			expectedFound: false,
			expectedProvider: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, exists := registry.Get(tt.providerName)
			if exists != tt.expectedFound {
				t.Errorf("Get(%q) exists = %v, expected %v", tt.providerName, exists, tt.expectedFound)
			}
			if exists && provider != tt.expectedProvider {
				t.Errorf("Get(%q) provider = %v, expected %v", tt.providerName, provider, tt.expectedProvider)
			}
		})
	}
}

func TestProviderRegistry_GetForModel(t *testing.T) {
	registry := NewProviderRegistry()
	
	provider1 := &MockSimpleProvider{
		name: "provider-1",
		supportedModels: map[string]bool{
			"model-1": true,
			"shared-model": true,
		},
	}
	
	provider2 := &MockSimpleProvider{
		name: "provider-2",
		supportedModels: map[string]bool{
			"model-2": true,
			"shared-model": true,
		},
	}
	
	provider3 := &MockSimpleProvider{
		name: "provider-3",
		supportedModels: map[string]bool{
			"model-3": true,
		},
	}
	
	registry.Register(provider1)
	registry.Register(provider2)
	registry.Register(provider3)
	
	tests := []struct {
		name            string
		model           string
		expectedFound   bool
		possibleProviders []string
	}{
		{
			name:            "Model supported by provider 1",
			model:           "model-1",
			expectedFound:   true,
			possibleProviders: []string{"provider-1"},
		},
		{
			name:            "Model supported by provider 2",
			model:           "model-2",
			expectedFound:   true,
			possibleProviders: []string{"provider-2"},
		},
		{
			name:            "Model supported by provider 3",
			model:           "model-3",
			expectedFound:   true,
			possibleProviders: []string{"provider-3"},
		},
		{
			name:            "Model supported by multiple providers",
			model:           "shared-model",
			expectedFound:   true,
			possibleProviders: []string{"provider-1", "provider-2"},
		},
		{
			name:            "Unsupported model",
			model:           "unsupported-model",
			expectedFound:   false,
			possibleProviders: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, exists := registry.GetForModel(tt.model)
			if exists != tt.expectedFound {
				t.Errorf("GetForModel(%q) exists = %v, expected %v", tt.model, exists, tt.expectedFound)
			}
			
			if exists {
				foundProviderName := provider.Name()
				
				// Check that the provider is one of the possible providers
				found := false
				for _, name := range tt.possibleProviders {
					if foundProviderName == name {
						found = true
						break
					}
				}
				
				if !found {
					t.Errorf("GetForModel(%q) returned provider %q, expected one of %v", tt.model, foundProviderName, tt.possibleProviders)
				}
			}
		})
	}
}

func TestProviderRegistry_All(t *testing.T) {
	registry := NewProviderRegistry()
	
	// Test with empty registry
	providers := registry.All()
	if len(providers) != 0 {
		t.Errorf("Expected empty registry to return empty slice, got %d providers", len(providers))
	}
	
	// Add some providers
	provider1 := &MockSimpleProvider{name: "provider-1", supportedModels: map[string]bool{"model-1": true}}
	provider2 := &MockSimpleProvider{name: "provider-2", supportedModels: map[string]bool{"model-2": true}}
	provider3 := &MockSimpleProvider{name: "provider-3", supportedModels: map[string]bool{"model-3": true}}
	
	registry.Register(provider1)
	registry.Register(provider2)
	registry.Register(provider3)
	
	// Check that all providers are returned
	providers = registry.All()
	if len(providers) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(providers))
	}
	
	// Check that all providers are in the result
	providerNames := make(map[string]bool)
	for _, p := range providers {
		providerNames[p.Name()] = true
	}
	
	expectedNames := []string{"provider-1", "provider-2", "provider-3"}
	for _, name := range expectedNames {
		if !providerNames[name] {
			t.Errorf("Expected provider %q to be in the result", name)
		}
	}
}

func TestProviderRegistry_ThreadSafety(t *testing.T) {
	registry := NewProviderRegistry()
	
	// Number of concurrent goroutines
	const numGoroutines = 10
	
	// Create a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // For both registering and getting providers
	
	// Channel to capture errors
	errChan := make(chan error, numGoroutines*2)
	
	// Concurrently register providers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			providerName := fmt.Sprintf("provider-%d", id)
			provider := &MockSimpleProvider{
				name: providerName,
				supportedModels: map[string]bool{
					fmt.Sprintf("model-%d", id): true,
				},
			}
			
			// Try to register the provider
			registry.Register(provider)
		}(i)
	}
	
	// Concurrently get providers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			providerName := fmt.Sprintf("provider-%d", id)
			modelName := fmt.Sprintf("model-%d", id)
			
			// Try to get the provider by name
			_, _ = registry.Get(providerName)
			
			// Try to get the provider by model
			_, _ = registry.GetForModel(modelName)
			
			// Try to get all providers
			_ = registry.All()
		}(i)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)
	
	// Check if there were any errors
	for err := range errChan {
		t.Errorf("Concurrent access error: %v", err)
	}
	
	// Check that all providers are registered
	providers := registry.All()
	if len(providers) != numGoroutines {
		t.Errorf("Expected %d providers after concurrent registration, got %d", numGoroutines, len(providers))
	}
}

func TestProvider_Interface(t *testing.T) {
	// Test that a provider implements all required methods of the Provider interface
	var _ Provider = &MockSimpleProvider{} // Will not compile if MockSimpleProvider doesn't implement Provider
	
	// Create a provider to test individual methods
	provider := &MockSimpleProvider{
		name: "test-provider",
		supportedModels: map[string]bool{
			"model-1": true,
		},
	}
	
	// Test Name method
	if name := provider.Name(); name != "test-provider" {
		t.Errorf("Name() = %q, expected %q", name, "test-provider")
	}
	
	// Test SupportsModel method
	if !provider.SupportsModel("model-1") {
		t.Errorf("SupportsModel(%q) = false, expected true", "model-1")
	}
	if provider.SupportsModel("unsupported-model") {
		t.Errorf("SupportsModel(%q) = true, expected false", "unsupported-model")
	}
	
	// Test CountTokens method
	_, err := provider.CountTokens(TokenCountParams{Model: "model-1"})
	if err != nil {
		t.Errorf("CountTokens() error = %v, expected nil", err)
	}
	
	// Test CalculatePrice method
	_, err = provider.CalculatePrice("model-1", 100, 50)
	if err != nil {
		t.Errorf("CalculatePrice() error = %v, expected nil", err)
	}
	
	// Test other methods (these are no-ops in the mock, so just ensure they don't panic)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetSDKClient() panicked: %v", r)
			}
		}()
		provider.SetSDKClient(nil)
	}()
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetModelInfo() panicked: %v", r)
			}
		}()
		_, _ = provider.GetModelInfo("model-1")
	}()
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ExtractTokenUsageFromResponse() panicked: %v", r)
			}
		}()
		_, _ = provider.ExtractTokenUsageFromResponse(nil)
	}()
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("UpdatePricing() panicked: %v", r)
			}
		}()
		_ = provider.UpdatePricing()
	}()
}
