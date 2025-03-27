package tokentracker

import (
	"encoding/json"
	"os"
	"sync"
	"time"
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
		Providers: map[string]ProviderConfig{
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

// EnableAutomaticPricingUpdates enables automatic pricing updates at the specified interval
func (c *Config) EnableAutomaticPricingUpdates(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.AutoUpdatePricing = true

	// Stop existing timer if any
	if c.pricingUpdateTimer != nil {
		c.pricingUpdateTimer.Stop()
	}

	// Create a new timer that will trigger pricing updates
	c.pricingUpdateTimer = time.AfterFunc(interval, func() {
		// This function will be called when the timer expires
		// It should trigger a pricing update and then reset the timer

		// Note: In a real implementation, this would call a method on TokenTracker
		// to update all pricing. Since we don't have direct access to TokenTracker here,
		// this is just a placeholder.

		// Reset the timer for the next interval
		c.pricingUpdateTimer.Reset(interval)
	})
}

// DisableAutomaticPricingUpdates disables automatic pricing updates
func (c *Config) DisableAutomaticPricingUpdates() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.AutoUpdatePricing = false

	// Stop the timer if it exists
	if c.pricingUpdateTimer != nil {
		c.pricingUpdateTimer.Stop()
		c.pricingUpdateTimer = nil
	}
}

// EnableUsageLogging enables logging of token usage to the specified file path
func (c *Config) EnableUsageLogging(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate that the path is writable
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	file.Close()

	c.UsageLogEnabled = true
	c.usageLogPath = path
	return nil
}

// DisableUsageLogging disables logging of token usage
func (c *Config) DisableUsageLogging() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.UsageLogEnabled = false
}

// GetUsageLogPath returns the path to the usage log file
func (c *Config) GetUsageLogPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.usageLogPath
}
