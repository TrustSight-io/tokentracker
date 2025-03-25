# Token Tracker Architecture

## Component Diagram

```mermaid
graph TD
    Client[Client Application] --> TokenTracker
    
    subgraph "Core Components"
        TokenTracker[TokenTracker Interface] --> DefaultTracker[DefaultTokenTracker]
        DefaultTracker --> Registry[ProviderRegistry]
        DefaultTracker --> Config[Configuration]
        Registry --> Provider[Provider Interface]
    end
    
    subgraph "Provider Implementations"
        Provider --> OpenAI[OpenAI Provider]
        Provider --> Gemini[Gemini Provider]
        Provider --> Claude[Claude Provider]
        
        OpenAI --> TikToken[tiktoken-go]
        Gemini --> GeminiLib[Gemini Library]
        Claude --> ClaudeLib[Claude Library]
    end
    
    subgraph "Supporting Components"
        Config --> Pricing[Model Pricing]
        DefaultTracker --> ErrorHandling[Error Handling]
        DefaultTracker --> Metrics[Usage Metrics]
    end
    
    style TokenTracker fill:#f9f,stroke:#333,stroke-width:2px
    style Provider fill:#bbf,stroke:#333,stroke-width:2px
    style Config fill:#bfb,stroke:#333,stroke-width:2px
```

## Sequence Diagram - Token Counting Flow

```mermaid
sequenceDiagram
    participant Client
    participant TokenTracker
    participant Registry
    participant Provider
    participant Tokenizer
    
    Client->>TokenTracker: CountTokens(params)
    TokenTracker->>Registry: GetProviderForModel(model)
    Registry-->>TokenTracker: provider
    TokenTracker->>Provider: CountTokens(params)
    Provider->>Tokenizer: Tokenize(text/messages)
    Tokenizer-->>Provider: tokenCount
    Provider-->>TokenTracker: TokenCount
    TokenTracker-->>Client: TokenCount
```

## Sequence Diagram - Usage Tracking Flow

```mermaid
sequenceDiagram
    participant Client
    participant TokenTracker
    participant Provider
    
    Client->>TokenTracker: TrackUsage(callParams, response)
    TokenTracker->>TokenTracker: CountTokens(callParams.Params)
    TokenTracker->>TokenTracker: ExtractResponseTokens(response)
    TokenTracker->>Provider: CalculatePrice(model, inputTokens, outputTokens)
    Provider-->>TokenTracker: price
    TokenTracker->>TokenTracker: CreateUsageMetrics()
    TokenTracker-->>Client: UsageMetrics
```

## Provider Registration Flow

```mermaid
graph LR
    A[Initialize TokenTracker] --> B[Create ProviderRegistry]
    B --> C[Register OpenAI Provider]
    C --> D[Register Gemini Provider]
    D --> E[Register Claude Provider]
    E --> F[Ready for Use]
    
    style A fill:#f9f,stroke:#333,stroke-width:1px
    style F fill:#bfb,stroke:#333,stroke-width:1px