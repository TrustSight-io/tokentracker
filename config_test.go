package tokentracker

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	// Check that the configuration is initialized with default values
	if config == nil {
		t.Fatal("NewConfig() returned nil")
	}

	// Check that providers are initialized
	if len(config.Providers) == 0 {
		t.Errorf("Expected providers to be initialized, got empty map")
	}

	// Check that specific providers exist
	providers := []string{"openai", "anthropic", "gemini"}
	for _, provider := range providers {
		if _, exists := config.Providers[provider]; !exists {
			t.Errorf("Expected provider %s to be initialized", provider)
		}
	}

	// Check that models are initialized for each provider
	for provider, providerConfig := range config.Providers {
		if len(providerConfig.Models) == 0 {
			t.Errorf("Expected models for provider %s to be initialized, got empty map", provider)
		}
	}
}

func TestConfig_GetModelPricing(t *testing.T) {
	config := NewConfig()

	tests := []struct {
		name           string
		provider       string
		model          string
		expectedExists bool
	}{
		{
			name:           "Valid OpenAI model",
			provider:       "openai",
			model:          "gpt-3.5-turbo",
			expectedExists: true,
		},
		{
			name:           "Valid Claude model",
			provider:       "anthropic",
			model:          "claude-3-haiku",
			expectedExists: true,
		},
		{
			name:           "Valid Gemini model",
			provider:       "gemini",
			model:          "gemini-pro",
			expectedExists: true,
		},
		{
			name:           "Invalid provider",
			provider:       "invalid-provider",
			model:          "gpt-3.5-turbo",
			expectedExists: false,
		},
		{
			name:           "Invalid model",
			provider:       "openai",
			model:          "invalid-model",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing, exists := config.GetModelPricing(tt.provider, tt.model)
			if exists != tt.expectedExists {
				t.Errorf("GetModelPricing() exists = %v, expected %v", exists, tt.expectedExists)
			}

			if exists {
				// Check that pricing has been initialized correctly
				if pricing.InputPricePerToken <= 0 {
					t.Errorf("Expected InputPricePerToken to be > 0, got %v", pricing.InputPricePerToken)
				}
				if pricing.OutputPricePerToken <= 0 {
					t.Errorf("Expected OutputPricePerToken to be > 0, got %v", pricing.OutputPricePerToken)
				}
				if pricing.Currency == "" {
					t.Errorf("Expected Currency to be non-empty")
				}
			}
		})
	}
}

func TestConfig_SetModelPricing(t *testing.T) {
	config := NewConfig()

	// Set custom pricing for a model
	customPricing := ModelPricing{
		InputPricePerToken:  0.0001,
		OutputPricePerToken: 0.0002,
		Currency:            "EUR",
	}

	// Test setting pricing for existing provider and model
	config.SetModelPricing("openai", "gpt-4", customPricing)
	pricing, exists := config.GetModelPricing("openai", "gpt-4")
	if !exists {
		t.Errorf("Expected pricing to exist after SetModelPricing()")
	}
	if pricing.InputPricePerToken != customPricing.InputPricePerToken {
		t.Errorf("Expected InputPricePerToken to be %v, got %v", customPricing.InputPricePerToken, pricing.InputPricePerToken)
	}
	if pricing.OutputPricePerToken != customPricing.OutputPricePerToken {
		t.Errorf("Expected OutputPricePerToken to be %v, got %v", customPricing.OutputPricePerToken, pricing.OutputPricePerToken)
	}
	if pricing.Currency != customPricing.Currency {
		t.Errorf("Expected Currency to be %v, got %v", customPricing.Currency, pricing.Currency)
	}

	// Test setting pricing for new provider and model
	config.SetModelPricing("new-provider", "new-model", customPricing)
	pricing, exists = config.GetModelPricing("new-provider", "new-model")
	if !exists {
		t.Errorf("Expected pricing to exist after SetModelPricing() for new provider and model")
	}
	if pricing.InputPricePerToken != customPricing.InputPricePerToken {
		t.Errorf("Expected InputPricePerToken to be %v, got %v", customPricing.InputPricePerToken, pricing.InputPricePerToken)
	}
}

func TestConfig_SaveAndLoadFromFile(t *testing.T) {
	config := NewConfig()

	// Set custom pricing
	customPricing := ModelPricing{
		InputPricePerToken:  0.0001,
		OutputPricePerToken: 0.0002,
		Currency:            "EUR",
	}
	config.SetModelPricing("test-provider", "test-model", customPricing)

	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "config-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up

	// Save the configuration to the file
	err = config.SaveToFile(tmpfile.Name())
	if err != nil {
		t.Errorf("SaveToFile() failed: %v", err)
	}

	// Create a new configuration
	newConfig := NewConfig()

	// Load the configuration from the file
	err = newConfig.LoadFromFile(tmpfile.Name())
	if err != nil {
		t.Errorf("LoadFromFile() failed: %v", err)
	}

	// Check that the custom pricing has been loaded correctly
	pricing, exists := newConfig.GetModelPricing("test-provider", "test-model")
	if !exists {
		t.Errorf("Expected pricing to exist after LoadFromFile()")
	}
	if pricing.InputPricePerToken != customPricing.InputPricePerToken {
		t.Errorf("Expected InputPricePerToken to be %v, got %v", customPricing.InputPricePerToken, pricing.InputPricePerToken)
	}
	if pricing.OutputPricePerToken != customPricing.OutputPricePerToken {
		t.Errorf("Expected OutputPricePerToken to be %v, got %v", customPricing.OutputPricePerToken, pricing.OutputPricePerToken)
	}
	if pricing.Currency != customPricing.Currency {
		t.Errorf("Expected Currency to be %v, got %v", customPricing.Currency, pricing.Currency)
	}
}

func TestConfig_AutomaticPricingUpdates(t *testing.T) {
	config := NewConfig()

	// Enable automatic pricing updates
	config.EnableAutomaticPricingUpdates(1 * time.Hour)

	// Check that automatic updates are enabled
	if !config.AutoUpdatePricing {
		t.Errorf("Expected AutoUpdatePricing to be true after EnableAutomaticPricingUpdates()")
	}

	// Disable automatic pricing updates
	config.DisableAutomaticPricingUpdates()

	// Check that automatic updates are disabled
	if config.AutoUpdatePricing {
		t.Errorf("Expected AutoUpdatePricing to be false after DisableAutomaticPricingUpdates()")
	}
}

func TestConfig_UsageLogging(t *testing.T) {
	config := NewConfig()

	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "usage-log-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up

	// Enable usage logging
	err = config.EnableUsageLogging(tmpfile.Name())
	if err != nil {
		t.Errorf("EnableUsageLogging() failed: %v", err)
	}

	// Check that usage logging is enabled
	if !config.UsageLogEnabled {
		t.Errorf("Expected UsageLogEnabled to be true after EnableUsageLogging()")
	}

	// Check that usage log path is set correctly
	if config.GetUsageLogPath() != tmpfile.Name() {
		t.Errorf("Expected usage log path to be %v, got %v", tmpfile.Name(), config.GetUsageLogPath())
	}

	// Disable usage logging
	config.DisableUsageLogging()

	// Check that usage logging is disabled
	if config.UsageLogEnabled {
		t.Errorf("Expected UsageLogEnabled to be false after DisableUsageLogging()")
	}

	// Test with non-existent directory
	err = config.EnableUsageLogging("/non-existent-directory/usage.log")
	if err == nil {
		t.Errorf("Expected EnableUsageLogging() to fail with non-existent directory")
	}
}
