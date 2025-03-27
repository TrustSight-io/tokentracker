# Token Tracker API Documentation

This document provides an overview of the Token Tracker API.

## Core Interfaces

### TokenTracker

The main interface that provides token counting and pricing functionality.

```go
type TokenTracker interface {
    // CountTokens counts tokens for a text string or chat messages
    CountTokens(params TokenCountParams) (TokenCount, error)

    // CalculatePrice calculates price based on token usage
    CalculatePrice(model string, inputTokens, outputTokens int) (Price, error)

    // TrackUsage tracks full usage for an LLM call
    TrackUsage(callParams CallParams, response interface{}) (UsageMetrics, error)

    // RegisterSDKClient registers an SDK client with the appropriate provider
    RegisterSDKClient(client SDKClient) error

    // UpdateAllPricing updates pricing information for all registered providers
    UpdateAllPricing() error

    // TrackTokenUsage extracts token usage from a provider response
    TrackTokenUsage(providerName string, response interface{}) (TokenCount, error)
}
```

### Provider

The interface for provider-specific implementations.

```go
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
```

### SDKClient

Interface for SDK clients that can be integrated with TokenTracker.

```go
type SDKClient interface {
    // GetProviderName returns the name of the LLM provider (e.g., "openai", "anthropic", "gemini")
    GetProviderName() string

    // GetClient returns the underlying SDK client instance
    GetClient() interface{}

    // GetSupportedModels returns a list of model identifiers supported by this provider
    GetSupportedModels() ([]string, error)

    // ExtractTokenUsageFromResponse extracts token usage information from an API response
    ExtractTokenUsageFromResponse(response interface{}) (common.TokenUsage, error)

    // FetchCurrentPricing fetches the current pricing information for all supported models
    FetchCurrentPricing() (map[string]common.ModelPricing, error)

    // UpdateProviderPricing updates the pricing information in the provider
    UpdateProviderPricing() error

    // TrackAPICall tracks an API call and returns usage metrics
    TrackAPICall(model string, response interface{}) (common.UsageMetrics, error)
}
```

## Key Types

### TokenCountParams

Parameters for token counting.

```go
type TokenCountParams struct {
    Model               string
    Text                *string
    Messages            []Message
    Tools               []Tool
    ToolChoice          *ToolChoice
    CountResponseTokens bool
}
```

### TokenCount

Result of token counting.

```go
type TokenCount struct {
    InputTokens    int
    ResponseTokens int
    TotalTokens    int
}
```

### Price

Pricing information for token usage.

```go
type Price struct {
    InputCost  float64
    OutputCost float64
    TotalCost  float64
    Currency   string
}
```

### UsageMetrics

Complete usage information for an LLM call.

```go
type UsageMetrics struct {
    TokenCount TokenCount
    Price      Price
    Duration   time.Duration
    Timestamp  time.Time
    Model      string
    Provider   string
}
```

### CallParams

Parameters for tracking an LLM call.

```go
type CallParams struct {
    Model     string
    Params    TokenCountParams
    StartTime time.Time
}
```

### ModelPricing

Pricing information for a specific model.

```go
type ModelPricing struct {
    InputPricePerToken  float64
    OutputPricePerToken float64
    Currency            string
}
```

## Configuration

The `Config` struct provides configuration options for the token tracker.

```go
type Config struct {
    Providers          map[string]ProviderConfig
    AutoUpdatePricing  bool
    UsageLogEnabled    bool
    // ... other fields
}
```

Key methods:

- `NewConfig()` - Creates a new configuration with default values
- `LoadFromFile(filename string)` - Loads configuration from a JSON file
- `SaveToFile(filename string)` - Saves configuration to a JSON file
- `GetModelPricing(provider, model string)` - Returns pricing information for a specific model
- `SetModelPricing(provider, model string, pricing ModelPricing)` - Sets pricing information for a specific model
- `EnableAutomaticPricingUpdates(interval time.Duration)` - Enables automatic pricing updates
- `DisableAutomaticPricingUpdates()` - Disables automatic pricing updates
- `EnableUsageLogging(path string)` - Enables logging of token usage
- `DisableUsageLogging()` - Disables logging of token usage

## Error Handling

Token Tracker provides structured error types for better error handling:

```go
type TokenTrackerError struct {
    Type    string
    Message string
    Cause   error
}
```

Common error types:

- `ErrInvalidModel` - Invalid model identifier
- `ErrInvalidParams` - Invalid parameters
- `ErrProviderNotFound` - Provider not found
- `ErrTokenizationFailed` - Tokenization failed
- `ErrPricingNotFound` - Pricing information not found
- `ErrPricingUpdateFailed` - Failed to update pricing information

## Using the API

### Basic Initialization

```go
config := tokentracker.NewConfig()
tracker := tokentracker.NewTokenTracker(config)

// Register providers
tracker.RegisterProvider(providers.NewOpenAIProvider(config))
tracker.RegisterProvider(providers.NewClaudeProvider(config))
tracker.RegisterProvider(providers.NewGeminiProvider(config))
```

### Token Counting

```go
params := tokentracker.TokenCountParams{
    Model: "gpt-3.5-turbo",
    Text:  &myText,
}

tokenCount, err := tracker.CountTokens(params)
```

### Price Calculation

```go
price, err := tracker.CalculatePrice("gpt-4", inputTokens, outputTokens)
```

### Tracking Usage

```go
usage, err := tracker.TrackUsage(callParams, response)
```

For comprehensive usage examples, refer to the [Getting Started Guide](getting_started.md) and the example applications in the repository.
