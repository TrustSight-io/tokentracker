package tokentracker

import (
	"encoding/json"
	"os"
	"sync"
)

// ModelPricing contains pricing information for a specific model
type ModelPricing struct {
	InputPricePerToken  float64
	OutputPricePerToken float64
	Currency            string
}

// ProviderConfig contains configuration for a specific provider
type ProviderConfig struct {
	Models map[string]ModelPricing
}

// Config contains the configuration for the token tracker
type Config struct {
	Providers          map[string]ProviderConfig
	AutoUpdatePricing  bool
	UsageLogEnabled    bool
	usageLogPath       string
	pricingUpdateTimer *time.Timer
	mu                 sync.RWMutex
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		Providers:         map[string]ProviderConfig{
			"openai": {
				Models: map[string]ModelPricing{
					"gpt-3.5-turbo": {
						InputPricePerToken:  0.0000015,
						OutputPricePerToken: 0.000002,
						Currency:            "USD",
					},
					"gpt-4": {
						InputPricePerToken:  0.00003,
						OutputPricePerToken: 0.00006,
						Currency:            "USD",
					},
				},
			},
			"anthropic": {
				Models: map[string]ModelPricing{
					"claude-3-haiku": {
						InputPricePerToken:  0.00000025,
						OutputPricePerToken: 0.00000125,
						Currency:            "USD",
					},
					"claude-3-sonnet": {
						InputPricePerToken:  0.000003,
						OutputPricePerToken: 0.000015,
						Currency:            "USD",
					},
					"claude-3-opus": {
						InputPricePerToken:  0.00001,
						OutputPricePerToken: 0.00003,
						Currency:            "USD",
					},
				},
			},
			"gemini": {
				Models: map[string]ModelPricing{
					"gemini-pro": {
						InputPricePerToken:  0.00000025,
						OutputPricePerToken: 0.0000005,
						Currency:            "USD",
					},
					"gemini-ultra": {
						InputPricePerToken:  0.00001,
						OutputPricePerToken: 0.00003,
						Currency:            "USD",
					},
				},
			},
		},
	}
}

// LoadFromFile loads configuration from a JSON file
func (c *Config) LoadFromFile(filename string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	c.Providers = config.Providers
	return nil
}

// SaveToFile saves configuration to a JSON file
func (c *Config) SaveToFile(filename string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetModelPricing returns pricing information for a specific model
func (c *Config) GetModelPricing(provider, model string) (ModelPricing, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	providerConfig, exists := c.Providers[provider]
	if !exists {
		return ModelPricing{}, false
	}

	pricing, exists := providerConfig.Models[model]
	return pricing, exists
}

// SetModelPricing sets pricing information for a specific model
func (c *Config) SetModelPricing(provider, model string, pricing ModelPricing) {
	c.mu.Lock()
	defer c.mu.Unlock()

	providerConfig, exists := c.Providers[provider]
	if !exists {
		providerConfig = ProviderConfig{
			Models: make(map[string]ModelPricing),
		}
		c.Providers[provider] = providerConfig
	}

	providerConfig.Models[model] = pricing
}
