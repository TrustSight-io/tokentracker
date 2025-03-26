package tokentracker

import "sync"

// Provider defines the interface for provider-specific implementations
type Provider interface {
	// Name returns the provider name
	Name() string

	// CountTokens counts tokens for the given parameters
	CountTokens(params TokenCountParams) (TokenCount, error)

	// CalculatePrice calculates price based on token usage
	CalculatePrice(model string, inputTokens, outputTokens int) (Price, error)

	// SupportsModel checks if the provider supports a specific model
	SupportsModel(model string) bool

	// SetSDKClient sets the provider-specific SDK client
	SetSDKClient(client interface{})

	// GetModelInfo returns information about a specific model
	GetModelInfo(model string) (interface{}, error)

	// ExtractTokenUsageFromResponse extracts token usage from a provider response
	ExtractTokenUsageFromResponse(response interface{}) (TokenCount, error)

	// UpdatePricing updates the pricing information for this provider
	UpdatePricing() error
}

// ProviderRegistry manages available providers
type ProviderRegistry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
func (r *ProviderRegistry) Register(provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Name()] = provider
}

// Get returns a provider by name
func (r *ProviderRegistry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, exists := r.providers[name]
	return provider, exists
}

// GetForModel returns a provider that supports the given model
func (r *ProviderRegistry) GetForModel(model string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, provider := range r.providers {
		if provider.SupportsModel(model) {
			return provider, true
		}
	}

	return nil, false
}

// All returns all registered providers
func (r *ProviderRegistry) All() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers
}
