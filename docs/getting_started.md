# Getting Started with Token Tracker

This guide will help you set up and use the Token Tracker library for tracking token usage and calculating pricing for API calls to various LLM providers.

## Installation

```bash
go get github.com/TrustSight-io/tokentracker
```

## Basic Setup

Here's how to set up the Token Tracker with the default configuration:

```go
import (
    "github.com/TrustSight-io/tokentracker"
    "github.com/TrustSight-io/tokentracker/providers"
)

// Create a new configuration
config := tokentracker.NewConfig()

// Create a new token tracker
tracker := tokentracker.NewTokenTracker(config)

// Register providers
tracker.RegisterProvider(providers.NewOpenAIProvider(config))
tracker.RegisterProvider(providers.NewGeminiProvider(config))
tracker.RegisterProvider(providers.NewClaudeProvider(config))
```

## Counting Tokens

### For Text

```go
text := "This is a sample text for token counting."
params := tokentracker.TokenCountParams{
    Model: "gpt-3.5-turbo",
    Text:  &text,
}

tokenCount, err := tracker.CountTokens(params)
if err != nil {
    // Handle error
}

fmt.Printf("Input tokens: %d\n", tokenCount.InputTokens)
fmt.Printf("Total tokens: %d\n", tokenCount.TotalTokens)
```

### For Chat Messages

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
    // Handle error
}

fmt.Printf("Input tokens: %d\n", tokenCount.InputTokens)
```

## Calculating Price

```go
price, err := tracker.CalculatePrice("gpt-4", 1000, 500)
if err != nil {
    // Handle error
}

fmt.Printf("Input cost: $%.6f\n", price.InputCost)
fmt.Printf("Output cost: $%.6f\n", price.OutputCost)
fmt.Printf("Total cost: $%.6f %s\n", price.TotalCost, price.Currency)
```

## Tracking Complete Usage

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
    // Handle error
}

fmt.Printf("Input tokens: %d\n", usage.TokenCount.InputTokens)
fmt.Printf("Response tokens: %d\n", usage.TokenCount.ResponseTokens)
fmt.Printf("Total tokens: %d\n", usage.TokenCount.TotalTokens)
fmt.Printf("Total cost: $%.6f %s\n", usage.Price.TotalCost, usage.Price.Currency)
fmt.Printf("Duration: %v\n", usage.Duration)
```

## Integration with SDK Clients

### Setting Up an SDK Wrapper

```go
import (
    "github.com/TrustSight-io/tokentracker/sdkwrappers"
)

// Create and register providers
openaiProvider := providers.NewOpenAIProvider(config)
tracker.RegisterProvider(openaiProvider)

// Create and register SDK wrapper
openaiWrapper := sdkwrappers.NewOpenAISDKWrapper("your-api-key", openaiProvider)
err := tracker.RegisterSDKClient(openaiWrapper)
if err != nil {
    // Handle error
}

// Use the SDK client directly
client := openaiWrapper.GetClient()
// Make API calls with the client
```

### Updating Pricing Information

```go
// Update pricing for all providers
err := tracker.UpdateAllPricing()
if err != nil {
    // Handle error
}

// Enable automatic pricing updates
config.EnableAutomaticPricingUpdates(24 * time.Hour)
```

### Logging Usage

```go
// Enable usage logging
err := config.EnableUsageLogging("token_usage.log")
if err != nil {
    // Handle error
}
```

## Advanced Configuration

### Custom Pricing

```go
// Update pricing for a specific model
config.SetModelPricing("openai", "gpt-4", tokentracker.ModelPricing{
    InputPricePerToken:  0.00003,
    OutputPricePerToken: 0.00006,
    Currency:            "USD",
})
```

### Saving and Loading Configuration

```go
// Save configuration to a file
err := config.SaveToFile("config.json")
if err != nil {
    // Handle error
}

// Load configuration from a file
err = config.LoadFromFile("config.json")
if err != nil {
    // Handle error
}
```

For more detailed examples, see the `example/main.go` and `cmd/main.go` files in the repository.
