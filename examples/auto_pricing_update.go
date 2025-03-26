package main

import (
	"fmt"
	"log"
	"time"

	"github.com/TrustSight-io/tokentracker"
	"github.com/TrustSight-io/tokentracker/providers"
	"github.com/TrustSight-io/tokentracker/sdkwrappers"
)

func main() {
	// Create a new configuration
	config := tokentracker.NewConfig()

	// Create a new token tracker
	tracker := tokentracker.NewTokenTracker(config)

	// Register providers
	claudeProvider := providers.NewClaudeProvider(config)
	tracker.RegisterProvider(claudeProvider)

	// Enable automatic pricing updates every 24 hours
	config.EnableAutomaticPricingUpdates(24 * time.Hour)
	fmt.Println("Automatic pricing updates enabled (every 24 hours)")

	// Enable usage logging
	err := config.EnableUsageLogging("token_usage.log")
	if err != nil {
		log.Fatalf("Failed to enable usage logging: %v", err)
	}
	fmt.Println("Usage logging enabled (token_usage.log)")

	// Create an Anthropic SDK wrapper
	// Note: In a real application, you would use your actual API key
	anthropicWrapper := sdkwrappers.NewAnthropicSDKWrapper("your-api-key-here")

	// Register the SDK client with the token tracker
	err = tracker.RegisterSDKClient(anthropicWrapper)
	if err != nil {
		log.Fatalf("Failed to register SDK client: %v", err)
	}
	fmt.Println("Registered Anthropic SDK client")

	// Update pricing information for all providers
	err = tracker.UpdateAllPricing()
	if err != nil {
		log.Fatalf("Failed to update pricing: %v", err)
	}
	fmt.Println("Updated pricing information for all providers")

	// In a real application, you would keep the program running
	// For this example, we'll just simulate a short run
	fmt.Println("Press Ctrl+C to exit...")
	
	// Keep the program running
	select {}
}
