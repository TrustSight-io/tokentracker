package main

import (
	"fmt"
	"time"

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

	// Example 1: Count tokens for a simple text
	text := "This is a sample text for token counting. It should give us a rough estimate of how many tokens are in this message."
	params := tokentracker.TokenCountParams{
		Model: "gpt-3.5-turbo",
		Text:  &text,
	}

	tokenCount, err := tracker.CountTokens(params)
	if err != nil {
		fmt.Printf("Error counting tokens: %v\n", err)
		return
	}

	fmt.Printf("Example 1 - Text token count for GPT-3.5-Turbo:\n")
	fmt.Printf("  Input tokens: %d\n", tokenCount.InputTokens)
	fmt.Printf("  Total tokens: %d\n\n", tokenCount.TotalTokens)

	// Example 2: Count tokens for chat messages
	messages := []tokentracker.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: "Tell me about token counting in large language models.",
		},
		{
			Role:    "assistant",
			Content: "Token counting is the process of estimating how many tokens (word pieces) are in a text. This is important for LLMs because they have context length limitations and pricing is often based on token usage.",
		},
	}

	chatParams := tokentracker.TokenCountParams{
		Model:              "claude-3-sonnet",
		Messages:           messages,
		CountResponseTokens: true,
	}

	chatTokenCount, err := tracker.CountTokens(chatParams)
	if err != nil {
		fmt.Printf("Error counting tokens for chat: %v\n", err)
		return
	}

	fmt.Printf("Example 2 - Chat token count for Claude-3-Sonnet:\n")
	fmt.Printf("  Input tokens: %d\n", chatTokenCount.InputTokens)
	fmt.Printf("  Response tokens (estimated): %d\n", chatTokenCount.ResponseTokens)
	fmt.Printf("  Total tokens: %d\n\n", chatTokenCount.TotalTokens)

	// Example 3: Calculate price
	price, err := tracker.CalculatePrice("gpt-4", 1000, 500)
	if err != nil {
		fmt.Printf("Error calculating price: %v\n", err)
		return
	}

	fmt.Printf("Example 3 - Price calculation for GPT-4:\n")
	fmt.Printf("  Input cost: $%.6f\n", price.InputCost)
	fmt.Printf("  Output cost: $%.6f\n", price.OutputCost)
	fmt.Printf("  Total cost: $%.6f %s\n\n", price.TotalCost, price.Currency)

	// Example 4: Track usage for a complete call
	callParams := tokentracker.CallParams{
		Model: "gemini-pro",
		Params: tokentracker.TokenCountParams{
			Model:    "gemini-pro",
			Messages: messages,
		},
		StartTime: time.Now().Add(-500 * time.Millisecond), // Simulate a call that took 500ms
	}

	// Simulate a response
	type MockResponse struct {
		Content string
	}
	response := MockResponse{
		Content: "This is a simulated response from the model.",
	}

	usage, err := tracker.TrackUsage(callParams, response)
	if err != nil {
		fmt.Printf("Error tracking usage: %v\n", err)
		return
	}

	fmt.Printf("Example 4 - Usage tracking for Gemini-Pro:\n")
	fmt.Printf("  Input tokens: %d\n", usage.TokenCount.InputTokens)
	fmt.Printf("  Response tokens: %d\n", usage.TokenCount.ResponseTokens)
	fmt.Printf("  Total tokens: %d\n", usage.TokenCount.TotalTokens)
	fmt.Printf("  Total cost: $%.6f %s\n", usage.Price.TotalCost, usage.Price.Currency)
	fmt.Printf("  Duration: %v\n", usage.Duration)
	fmt.Printf("  Provider: %s\n", usage.Provider)
}