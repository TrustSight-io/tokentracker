# Token Tracker

A Golang module for tracking token usage and calculating pricing for API calls to various LLM providers (Gemini, Claude, OpenAI).

## Features

- Token counting for text and chat messages
- Support for multiple LLM providers:
  - OpenAI (GPT-3.5, GPT-4)
  - Anthropic (Claude 3 Haiku, Sonnet, Opus)
  - Google (Gemini Pro, Ultra)
- Price calculation based on model-specific pricing
- Usage tracking for complete LLM calls
- Configurable pricing and model settings
- Thread-safe implementation

## Installation

```bash
go get github.com/TrustSight-io/tokentracker
```

## Usage

### Basic Token Counting

```go
package main

import (
	"fmt"

	"github.com/TrustSight-io/tokentracker"
	"github.com/TrustSight-io/tokentracker/providers"
)

func main() {
	// Create a new configuration
	config := tokentracker.NewConfig()

	// Create a new token tracker
	tracker := tokentracker.NewTokenTracker(config)

	// Register providers
	tracker.RegisterProvider(providers.NewOpenAIProvider(config))
	tracker.RegisterProvider(providers.NewGeminiProvider(config))
	tracker.RegisterProvider(providers.NewClaudeProvider(config))

	// Count tokens for a text
	text := "This is a sample text for token counting."
	params := tokentracker.TokenCountParams{
		Model: "gpt-3.5-turbo",
		Text:  &text,
	}

	tokenCount, err := tracker.CountTokens(params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Input tokens: %d\n", tokenCount.InputTokens)
	fmt.Printf("Total tokens: %d\n", tokenCount.TotalTokens)
}
```

### Counting Tokens for Chat Messages

```go
messages := []tokentracker.Message{
	{
		Role:    "system",
		Content: "You are a helpful assistant.",
	},
	{
		Role:    "user",
		Content: "Tell me about token counting.",
	},
}

params := tokentracker.TokenCountParams{
	Model:    "gpt-4",
	Messages: messages,
}

tokenCount, err := tracker.CountTokens(params)
if err != nil {
	fmt.Printf("Error: %v\n", err)
	return
}

fmt.Printf("Input tokens: %d\n", tokenCount.InputTokens)
```

### Calculating Price

```go
price, err := tracker.CalculatePrice("gpt-4", 1000, 500)
if err != nil {
	fmt.Printf("Error: %v\n", err)
	return
}

fmt.Printf("Input cost: $%.6f\n", price.InputCost)
fmt.Printf("Output cost: $%.6f\n", price.OutputCost)
fmt.Printf("Total cost: $%.6f %s\n", price.TotalCost, price.Currency)
```

### Tracking Complete Usage

```go
import "time"

callParams := tokentracker.CallParams{
	Model: "claude-3-haiku",
	Params: tokentracker.TokenCountParams{
		Model:    "claude-3-haiku",
		Messages: messages,
	},
	StartTime: time.Now(),
}

// Make your API call here
// ...

// Track usage
usage, err := tracker.TrackUsage(callParams, response)
if err != nil {
	fmt.Printf("Error: %v\n", err)
	return
}

fmt.Printf("Input tokens: %d\n", usage.TokenCount.InputTokens)
fmt.Printf("Response tokens: %d\n", usage.TokenCount.ResponseTokens)
fmt.Printf("Total tokens: %d\n", usage.TokenCount.TotalTokens)
fmt.Printf("Total cost: $%.6f %s\n", usage.Price.TotalCost, usage.Price.Currency)
fmt.Printf("Duration: %v\n", usage.Duration)
```

## Configuration

The token tracker comes with default pricing for common models, but you can customize it:

```go
config := tokentracker.NewConfig()

// Update pricing for a specific model
config.SetModelPricing("openai", "gpt-4", tokentracker.ModelPricing{
	InputPricePerToken:  0.00003,
	OutputPricePerToken: 0.00006,
	Currency:            "USD",
})

// Save configuration to a file
err := config.SaveToFile("config.json")
if err != nil {
	fmt.Printf("Error saving config: %v\n", err)
}

// Load configuration from a file
err = config.LoadFromFile("config.json")
if err != nil {
	fmt.Printf("Error loading config: %v\n", err)
}
```

## Limitations

- The token counting for Gemini and Claude models uses approximations and should be replaced with official tokenizers when available.
- Image token counting is simplified and may not be accurate for all use cases.
- Tool calls token counting is approximate and may need adjustments based on actual usage.

## SDK Integration

The token tracker provides integration with official LLM SDK clients through the `SDKClientWrapper` interface. This allows you to easily track token usage and costs for API calls made with official SDKs.

### Registering SDK Clients

```go
// Create a new configuration and token tracker
config := tokentracker.NewConfig()
tracker := tokentracker.NewTokenTracker(config)

// Register providers
claudeProvider := providers.NewClaudeProvider(config)
tracker.RegisterProvider(claudeProvider)

// Create an Anthropic SDK wrapper with your API key
anthropicWrapper := sdkwrappers.NewAnthropicSDKWrapper("your-api-key", claudeProvider)

// Register the SDK client with the token tracker
err := tracker.RegisterSDKClient(anthropicWrapper)
if err != nil {
    log.Fatalf("Failed to register SDK client: %v", err)
}
```

### Updating Pricing Information

```go
// Update pricing information for all providers
err := tracker.UpdateAllPricing()
if err != nil {
    log.Fatalf("Failed to update pricing: %v", err)
}

// Enable automatic pricing updates every 24 hours
config.EnableAutomaticPricingUpdates(24 * time.Hour)
```

### Tracking Usage from API Responses

```go
// Make an API call using the SDK client
client := anthropicWrapper.GetClient()
resp, err := client.Messages.Create(context.Background(), &anthropic.MessageRequest{
    Model: sdkwrappers.ClaudeHaiku,
    MaxTokens: 1000,
    Messages: []anthropic.Message{
        {
            Role: "user",
            Content: "Explain token counting in LLMs.",
        },
    },
})
if err != nil {
    log.Fatalf("API call failed: %v", err)
}

// Track token usage from the response
tokenCount, err := tracker.TrackTokenUsage("anthropic", resp)
if err != nil {
    log.Fatalf("Failed to track token usage: %v", err)
}

fmt.Printf("Input tokens: %d\n", tokenCount.InputTokens)
fmt.Printf("Response tokens: %d\n", tokenCount.ResponseTokens)
fmt.Printf("Total tokens: %d\n", tokenCount.TotalTokens)

// Get detailed usage metrics including pricing
metrics, err := anthropicWrapper.TrackAPICall(sdkwrappers.ClaudeHaiku, resp)
if err != nil {
    log.Fatalf("Failed to track API call: %v", err)
}

fmt.Printf("Total cost: $%.6f %s\n", metrics.Price.TotalCost, metrics.Price.Currency)
```

### Example Usage with OpenAI

```go
// Register OpenAI provider
openaiProvider := providers.NewOpenAIProvider(config)
tracker.RegisterProvider(openaiProvider)

// Create and register OpenAI SDK wrapper
openaiWrapper := sdkwrappers.NewOpenAISDKWrapper("your-openai-api-key", openaiProvider)
tracker.RegisterSDKClient(openaiWrapper)

// Enable usage logging
config.EnableUsageLogging("token_usage.log")

// Make API calls and track usage
// ...
```

## License

MIT
