# Contributing to Token Tracker

We love your input! We want to make contributing to Token Tracker as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

### Pull Requests

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

### Testing

Before submitting a PR, make sure all tests pass:

```bash
make test
```

To run the examples:

```bash
make example-original
make example
```

### Coding Style

Please follow the existing code style in the project. Some key points:

- Use meaningful variable and function names
- Add proper documentation for public functions and types
- Keep functions small and focused on a single task
- Use error handling patterns consistent with the rest of the codebase

## Project Structure

- `/` - Root package containing core functionality
- `/common` - Common types shared across packages
- `/providers` - Provider-specific implementations
- `/sdkwrappers` - SDK client wrappers for integration
- `/example` - Example application
- `/cmd` - Command-line application
- `/docs` - Documentation

## Adding Support for New Providers

To add support for a new LLM provider:

1. Create a new file in the `providers` directory (e.g., `providers/newprovider.go`)
2. Implement the `Provider` interface defined in `provider.go`
3. Create a constructor function (e.g., `NewMyProvider`) that initializes the provider
4. Implement token counting logic specific to that provider
5. Add pricing information to the default configuration
6. Create tests for the new provider

Example skeleton for a new provider:

```go
package providers

import (
    "sync"
    
    "github.com/TrustSight-io/tokentracker"
)

// MyProvider implements the Provider interface for MyModel models
type MyProvider struct {
    config    *tokentracker.Config
    sdkClient interface{}
    modelInfo map[string]interface{}
    mu        sync.RWMutex
}

// NewMyProvider creates a new MyModel provider
func NewMyProvider(config *tokentracker.Config) *MyProvider {
    provider := &MyProvider{
        config:    config,
        modelInfo: make(map[string]interface{}),
    }
    
    // Initialize with default model info
    provider.initializeModelInfo()
    
    return provider
}

// Name returns the provider name
func (p *MyProvider) Name() string {
    return "myprovider"
}

// Implement other required methods...
```

## Adding SDK Integration

To add SDK integration for a new provider:

1. Create a new file in the `sdkwrappers` directory (e.g., `sdkwrappers/myprovider.go`)
2. Implement the `SDKClientWrapper` interface defined in `sdk.go`
3. Create appropriate extraction and tracking methods

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License.
