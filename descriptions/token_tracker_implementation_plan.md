# Golang Token Usage Tracking Module - Implementation Plan

## Project Overview

This document outlines the implementation plan for a modular Golang Token Usage Tracking Module that can accurately track token usage and calculate pricing for API calls to various LLM providers (Gemini, Claude, OpenAI).

## Project Structure

```
tokentracker/
├── tokentracker.go       # Main package, core interfaces
├── config.go             # Configuration system
├── models.go             # Common data structures and types
├── provider.go           # Provider interface and registry
├── providers/
│   ├── gemini.go         # Gemini implementation
│   ├── claude.go         # Claude implementation
│   ├── openai.go         # OpenAI implementation
├── errors.go             # Custom error types
├── utils.go              # Utility functions
```

## Implementation Phases

### Phase 1: Core Framework

1. **Define Core Interfaces and Types**
   - Implement the `TokenTracker` interface as specified
   - Create supporting types (`TokenCountParams`, `TokenCount`, `Price`, `UsageMetrics`)
   - Define the `Provider` interface for provider-specific implementations
   - Implement a provider registry for dynamic selection

2. **Configuration System**
   - Create a flexible configuration system for model pricing
   - Support environment variables and file-based configuration
   - Implement configuration validation and defaults
   - Allow for runtime updates to pricing

3. **Error Handling and Utilities**
   - Define custom error types for different failure scenarios
   - Implement utility functions for common operations
   - Add logging helpers for tracking usage

### Phase 2: Provider Implementations

1. **Provider Interface**
   - Define common provider behavior
   - Create factory methods for provider instantiation
   - Implement provider-specific configuration

2. **OpenAI Implementation**
   - Use a Go port of tiktoken for accurate token counting
   - Support different GPT models (3.5, 4)
   - Implement message and tool call counting
   - Add pricing calculation

3. **Gemini Implementation**
   - Integrate with Gemini's token counting approach
   - Support all Gemini models
   - Implement message and tool call counting
   - Add pricing calculation

4. **Claude Implementation**
   - Implement Anthropic's token counting logic
   - Support Claude 3 models (Haiku, Sonnet, Opus)
   - Implement message and tool call counting
   - Add pricing calculation

## Detailed Implementation

### Core Types and Interfaces (models.go)

```go
// models.go
package tokentracker

import "time"

// Message represents a chat message
type Message struct {
    Role    string      `json:"role"`
    Content interface{} `json:"content"` // string or ContentPart array
}

// ContentPart represents a part of a message content (text or image)
type ContentPart struct {
    Type  string      `json:"type"`
    Text  string      `json:"text,omitempty"`
    Image interface{} `json:"image,omitempty"`
}

// Tool represents a function or tool definition
type Tool struct {
    Type     string      `json:"type"`
    Function interface{} `json:"function,omitempty"`
}

// ToolChoice represents a tool choice specification
type ToolChoice struct {
    Type     string      `json:"type,omitempty"`
    Function interface{} `json:"function,omitempty"`
}

// TokenCountParams contains parameters for token counting
type TokenCountParams struct {
    Model              string
    Text               *string
    Messages           []Message
    Tools              []Tool
    ToolChoice         *ToolChoice
    CountResponseTokens bool
}

// TokenCount contains token counting results
type TokenCount struct {
    InputTokens    int
    ResponseTokens int
    TotalTokens    int
}

// Price contains pricing information
type Price struct {
    InputCost  float64
    OutputCost float64
    TotalCost  float64
    Currency   string
}

// UsageMetrics contains complete usage information
type UsageMetrics struct {
    TokenCount TokenCount
    Price      Price
    Duration   time.Duration
    Timestamp  time.Time
    Model      string
    Provider   string
}

// CallParams contains parameters for an LLM call
type CallParams struct {
    Model     string
    Params    TokenCountParams
    StartTime time.Time
}
```

### Provider Interface (provider.go)

```go
// provider.go
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
```

### Main TokenTracker Implementation (tokentracker.go)

```go
// tokentracker.go
package tokentracker

import (
    "errors"
    "time"
)

// TokenTracker interface defines the main functionality
type TokenTracker interface {
    // CountTokens counts tokens for a text string or chat messages
    CountTokens(params TokenCountParams) (TokenCount, error)
    
    // CalculatePrice calculates price based on token usage
    CalculatePrice(model string, inputTokens, outputTokens int) (Price, error)
    
    // TrackUsage tracks full usage for an LLM call
    TrackUsage(callParams CallParams, response interface{}) (UsageMetrics, error)
}

// DefaultTokenTracker implements the TokenTracker interface
type DefaultTokenTracker struct {
    registry *ProviderRegistry
    config   *Config
}

// NewTokenTracker creates a new token tracker with the given configuration
func NewTokenTracker(config *Config) *DefaultTokenTracker {
    registry := NewProviderRegistry()
    
    // Register default providers here or allow caller to register them
    
    return &DefaultTokenTracker{
        registry: registry,
        config:   config,
    }
}

// RegisterProvider registers a provider with the token tracker
func (t *DefaultTokenTracker) RegisterProvider(provider Provider) {
    t.registry.Register(provider)
}

// CountTokens counts tokens for the given parameters
func (t *DefaultTokenTracker) CountTokens(params TokenCountParams) (TokenCount, error) {
    if params.Model == "" {
        return TokenCount{}, errors.New("model is required")
    }
    
    provider, exists := t.registry.GetForModel(params.Model)
    if !exists {
        return TokenCount{}, errors.New("no provider found for model: " + params.Model)
    }
    
    return provider.CountTokens(params)
}

// CalculatePrice calculates price based on token usage
func (t *DefaultTokenTracker) CalculatePrice(model string, inputTokens, outputTokens int) (Price, error) {
    if model == "" {
        return Price{}, errors.New("model is required")
    }
    
    provider, exists := t.registry.GetForModel(model)
    if !exists {
        return Price{}, errors.New("no provider found for model: " + model)
    }
    
    return provider.CalculatePrice(model, inputTokens, outputTokens)
}

// TrackUsage tracks full usage for an LLM call
func (t *DefaultTokenTracker) TrackUsage(callParams CallParams, response interface{}) (UsageMetrics, error) {
    // Get input token count
    inputCount, err := t.CountTokens(callParams.Params)
    if err != nil {
        return UsageMetrics{}, err
    }
    
    // Extract response tokens from the response object
    // This will be provider-specific and depend on the response structure
    var outputTokens int
    
    // Calculate price
    price, err := t.CalculatePrice(callParams.Model, inputCount.InputTokens, outputTokens)
    if err != nil {
        return UsageMetrics{}, err
    }
    
    // Calculate duration
    duration := time.Since(callParams.StartTime)
    
    // Get provider name
    provider, _ := t.registry.GetForModel(callParams.Model)
    providerName := provider.Name()
    
    // Create usage metrics
    metrics := UsageMetrics{
        TokenCount: TokenCount{
            InputTokens:    inputCount.InputTokens,
            ResponseTokens: outputTokens,
            TotalTokens:    inputCount.InputTokens + outputTokens,
        },
        Price:     price,
        Duration:  duration,
        Timestamp: time.Now(),
        Model:     callParams.Model,
        Provider:  providerName,
    }
    
    return metrics, nil
}
```

### Configuration System (config.go)

```go
// config.go
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
    Providers map[string]ProviderConfig
    mu        sync.RWMutex
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
```

### Error Handling (errors.go)

```go
// errors.go
package tokentracker

import "fmt"

// Error types
const (
    ErrInvalidModel      = "invalid_model"
    ErrInvalidParams     = "invalid_params"
    ErrProviderNotFound  = "provider_not_found"
    ErrTokenizationFailed = "tokenization_failed"
    ErrPricingNotFound   = "pricing_not_found"
)

// TokenTrackerError represents an error in the token tracker
type TokenTrackerError struct {
    Type    string
    Message string
    Cause   error
}

// Error returns the error message
func (e *TokenTrackerError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *TokenTrackerError) Unwrap() error {
    return e.Cause
}

// NewError creates a new TokenTrackerError
func NewError(errType, message string, cause error) *TokenTrackerError {
    return &TokenTrackerError{
        Type:    errType,
        Message: message,
        Cause:   cause,
    }
}
```

### Provider Implementation Example (OpenAI)

```go
// providers/openai.go
package providers

import (
    "errors"
    
    "github.com/yourusername/tokentracker"
    "github.com/pkoukk/tiktoken-go"
)

// OpenAIProvider implements the Provider interface for OpenAI models
type OpenAIProvider struct {
    config *tokentracker.Config
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *tokentracker.Config) *OpenAIProvider {
    return &OpenAIProvider{
        config: config,
    }
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
    return "openai"
}

// SupportsModel checks if the provider supports a specific model
func (p *OpenAIProvider) SupportsModel(model string) bool {
    supportedModels := map[string]bool{
        "gpt-3.5-turbo": true,
        "gpt-4":         true,
        "gpt-4-turbo":   true,
        // Add more models as needed
    }
    
    return supportedModels[model]
}

// CountTokens counts tokens for the given parameters
func (p *OpenAIProvider) CountTokens(params tokentracker.TokenCountParams) (tokentracker.TokenCount, error) {
    if params.Model == "" {
        return tokentracker.TokenCount{}, errors.New("model is required")
    }
    
    // Get the encoding for the model
    encoding, err := p.getEncoding(params.Model)
    if err != nil {
        return tokentracker.TokenCount{}, err
    }
    
    var inputTokens int
    
    // Count tokens based on input type
    if params.Text != nil {
        // Count tokens for text
        inputTokens = len(encoding.Encode(*params.Text, nil, nil))
    } else if len(params.Messages) > 0 {
        // Count tokens for messages
        inputTokens, err = p.countMessageTokens(encoding, params.Messages, params.Tools, params.ToolChoice)
        if err != nil {
            return tokentracker.TokenCount{}, err
        }
    } else {
        return tokentracker.TokenCount{}, errors.New("either text or messages must be provided")
    }
    
    // Estimate response tokens if requested
    var responseTokens int
    if params.CountResponseTokens {
        // This is a simplified estimation
        responseTokens = p.estimateResponseTokens(params.Model, inputTokens)
    }
    
    return tokentracker.TokenCount{
        InputTokens:    inputTokens,
        ResponseTokens: responseTokens,
        TotalTokens:    inputTokens + responseTokens,
    }, nil
}

// CalculatePrice calculates price based on token usage
func (p *OpenAIProvider) CalculatePrice(model string, inputTokens, outputTokens int) (tokentracker.Price, error) {
    pricing, exists := p.config.GetModelPricing("openai", model)
    if !exists {
        return tokentracker.Price{}, errors.New("pricing not found for model: " + model)
    }
    
    inputCost := float64(inputTokens) * pricing.InputPricePerToken
    outputCost := float64(outputTokens) * pricing.OutputPricePerToken
    
    return tokentracker.Price{
        InputCost:  inputCost,
        OutputCost: outputCost,
        TotalCost:  inputCost + outputCost,
        Currency:   pricing.Currency,
    }, nil
}

// getEncoding returns the tiktoken encoding for the given model
func (p *OpenAIProvider) getEncoding(model string) (*tiktoken.Tiktoken, error) {
    var encodingName string
    
    // Determine the encoding name based on the model
    switch model {
    case "gpt-3.5-turbo", "gpt-4", "gpt-4-turbo":
        encodingName = "cl100k_base"
    default:
        return nil, errors.New("unsupported model: " + model)
    }
    
    encoding, err := tiktoken.GetEncoding(encodingName)
    if err != nil {
        return nil, err
    }
    
    return encoding, nil
}

// countMessageTokens counts tokens for chat messages
func (p *OpenAIProvider) countMessageTokens(encoding *tiktoken.Tiktoken, messages []tokentracker.Message, tools []tokentracker.Tool, toolChoice *tokentracker.ToolChoice) (int, error) {
    // Implementation based on OpenAI's token counting logic
    // This is a simplified version and would need to be expanded
    
    // Base tokens for the messages format
    tokens := 3 // Every reply is primed with <|start|>assistant<|message|>
    
    // Count tokens for each message
    for _, message := range messages {
        // Add tokens for message role
        tokens += 4 // Every message follows <|start|>{role}<|message|>
        
        // Count tokens for content
        switch content := message.Content.(type) {
        case string:
            tokens += len(encoding.Encode(content, nil, nil))
        case []tokentracker.ContentPart:
            for _, part := range content {
                if part.Type == "text" {
                    tokens += len(encoding.Encode(part.Text, nil, nil))
                } else if part.Type == "image" {
                    // Simplified image token counting
                    tokens += 1000 // Placeholder value
                }
            }
        }
    }
    
    // Count tokens for tools if provided
    if len(tools) > 0 {
        // Simplified tool token counting
        for _, tool := range tools {
            // Add base tokens for tool
            tokens += 4
            
            // Count tokens for function definition
            // This would need to be expanded based on the actual structure
            tokens += 50 // Placeholder value
        }
    }
    
    // Count tokens for tool choice if provided
    if toolChoice != nil {
        tokens += 10 // Placeholder value
    }
    
    return tokens, nil
}

// estimateResponseTokens estimates the number of response tokens
func (p *OpenAIProvider) estimateResponseTokens(model string, inputTokens int) int {
    // This is a very simplified estimation
    // In a real implementation, this would be more sophisticated
    
    switch model {
    case "gpt-3.5-turbo":
        return inputTokens / 2
    case "gpt-4", "gpt-4-turbo":
        return inputTokens
    default:
        return inputTokens / 2
    }
}
```

## Implementation Strategy

1. **Start with Core Framework**
   - Implement the basic interfaces and types
   - Create the configuration system
   - Set up the provider registry

2. **Implement OpenAI Provider First**
   - OpenAI has the most established tokenization approach
   - Use tiktoken-go for accurate token counting
   - This will serve as a template for other providers

3. **Add Gemini Provider**
   - Research Gemini's token counting approach
   - Implement based on Google's generative AI Go library

4. **Add Claude Provider**
   - Research Claude's token counting approach
   - Implement based on Anthropic's documentation

5. **Finalize and Test**
   - Add comprehensive error handling
   - Implement performance optimizations
   - Create usage examples

## Dependencies

1. **Required Dependencies**
   - `github.com/pkoukk/tiktoken-go` for OpenAI token counting
   - Potentially a Go library for Gemini API
   - Potentially a Go library for Claude API

2. **Optional Dependencies**
   - A logging library (e.g., `go.uber.org/zap`)
   - A configuration library (e.g., `github.com/spf13/viper`)

## Testing Strategy

1. **Unit Tests**
   - Test each provider implementation
   - Test token counting for different input types
   - Test price calculation

2. **Integration Tests**
   - Test the full flow from token counting to price calculation
   - Test provider selection logic
   - Test configuration changes

3. **Benchmarks**
   - Measure performance of token counting
   - Compare different implementation approaches

## Implementation Timeline

1. **Phase 1 (Core Framework)**: 1-2 days
2. **Phase 2 (Provider Implementations)**: 2-3 days
3. **Testing and Optimization**: 1-2 days

## Next Steps

1. Set up the project structure
2. Implement the core interfaces and types
3. Create the configuration system
4. Implement the OpenAI provider as a starting point