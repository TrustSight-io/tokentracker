package main

import (
	"fmt"
	"os"
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
	openaiProvider := providers.NewOpenAIProvider(config)
	claudeProvider := providers.NewClaudeProvider(config)
	geminiProvider := providers.NewGeminiProvider(config)

	tracker.RegisterProvider(openaiProvider)
	tracker.RegisterProvider(claudeProvider)
	tracker.RegisterProvider(geminiProvider)

	// Enable automatic pricing updates (every 24 hours)
	config.EnableAutomaticPricingUpdates(24 * time.Hour)

	// Enable usage logging
	logPath := "token_usage.log"
	if err := config.EnableUsageLogging(logPath); err != nil {
		fmt.Printf("Warning: Unable to enable usage logging: %v\n", err)
	}

	// Basic token counting for text
	demoBasicTokenCounting(tracker)

	// Token counting for chat messages
	demoChatTokenCounting(tracker)

	// Price calculation
	demoPriceCalculation(tracker)

	// Tracking usage
	demoUsageTracking(tracker)

	// SDK client integration (if API keys are available)
	demoSDKIntegration(tracker, openaiProvider, claudeProvider, geminiProvider)
}

func demoBasicTokenCounting(tracker *tokentracker.DefaultTokenTracker) {
	fmt.Println("=== Basic Token Counting ===")

	text := "This is a sample text for token counting. It would be used to demonstrate the token counting functionality."
	params := tokentracker.TokenCountParams{
		Model: "gpt-3.5-turbo",
		Text:  &text,
	}

	tokenCount, err := tracker.CountTokens(params)
	if err != nil {
		fmt.Printf("Error counting tokens: %v\n", err)
		return
	}

	fmt.Printf("Text: %s\n", text)
	fmt.Printf("Model: %s\n", params.Model)
	fmt.Printf("Input tokens: %d\n", tokenCount.InputTokens)
	fmt.Printf("Total tokens: %d\n\n", tokenCount.TotalTokens)
}

func demoChatTokenCounting(tracker *tokentracker.DefaultTokenTracker) {
	fmt.Println("=== Chat Token Counting ===")

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

	models := []string{"gpt-4", "claude-3-sonnet", "gemini-pro"}

	for _, model := range models {
		params := tokentracker.TokenCountParams{
			Model:               model,
			Messages:            messages,
			CountResponseTokens: true,
		}

		tokenCount, err := tracker.CountTokens(params)
		if err != nil {
			fmt.Printf("Error counting tokens for %s: %v\n", model, err)
			continue
		}

		fmt.Printf("Model: %s\n", model)
		fmt.Printf("Input tokens: %d\n", tokenCount.InputTokens)
		fmt.Printf("Response tokens (estimated): %d\n", tokenCount.ResponseTokens)
		fmt.Printf("Total tokens: %d\n\n", tokenCount.TotalTokens)
	}
}

func demoPriceCalculation(tracker *tokentracker.DefaultTokenTracker) {
	fmt.Println("=== Price Calculation ===")

	modelPricings := []struct {
		model        string
		inputTokens  int
		outputTokens int
	}{
		{"gpt-3.5-turbo", 1000, 500},
		{"gpt-4", 1000, 500},
		{"claude-3-haiku", 1000, 500},
		{"claude-3-sonnet", 1000, 500},
		{"claude-3-opus", 1000, 500},
		{"gemini-pro", 1000, 500},
		{"gemini-ultra", 1000, 500},
	}

	for _, mp := range modelPricings {
		price, err := tracker.CalculatePrice(mp.model, mp.inputTokens, mp.outputTokens)
		if err != nil {
			fmt.Printf("Error calculating price for %s: %v\n", mp.model, err)
			continue
		}

		fmt.Printf("Model: %s\n", mp.model)
		fmt.Printf("Input tokens: %d, Output tokens: %d\n", mp.inputTokens, mp.outputTokens)
		fmt.Printf("Input cost: $%.6f\n", price.InputCost)
		fmt.Printf("Output cost: $%.6f\n", price.OutputCost)
		fmt.Printf("Total cost: $%.6f %s\n\n", price.TotalCost, price.Currency)
	}
}

func demoUsageTracking(tracker *tokentracker.DefaultTokenTracker) {
	fmt.Println("=== Usage Tracking ===")

	// Create messages for the call
	messages := []tokentracker.Message{
		{
			Role:    "user",
			Content: "Explain how tokens work in large language models.",
		},
	}

	// Create call parameters
	callParams := tokentracker.CallParams{
		Model: "gpt-3.5-turbo",
		Params: tokentracker.TokenCountParams{
			Model:    "gpt-3.5-turbo",
			Messages: messages,
		},
		StartTime: time.Now().Add(-300 * time.Millisecond), // Simulate a call that took 300ms
	}

	// Simulate a response
	type MockResponse struct {
		Content string
	}

	response := MockResponse{
		Content: "Tokens are the basic units of text that language models process. They are not exactly words, but rather word pieces or subwords. For example, the word 'tokenization' might be split into the tokens 'token' and 'ization'. Most modern language models use subword tokenization algorithms like BPE (Byte Pair Encoding) or WordPiece. This approach allows models to handle a large vocabulary while keeping the token set manageable. Each token is assigned a unique ID that the model uses for processing. The number of tokens in a text affects both the model's processing time and the cost of API calls, as providers typically charge based on the number of tokens processed.",
	}

	// Track usage
	usage, err := tracker.TrackUsage(callParams, response)
	if err != nil {
		fmt.Printf("Error tracking usage: %v\n", err)
		return
	}

	fmt.Printf("Model: %s\n", callParams.Model)
	fmt.Printf("Input tokens: %d\n", usage.TokenCount.InputTokens)
	fmt.Printf("Response tokens: %d\n", usage.TokenCount.ResponseTokens)
	fmt.Printf("Total tokens: %d\n", usage.TokenCount.TotalTokens)
	fmt.Printf("Total cost: $%.6f %s\n", usage.Price.TotalCost, usage.Price.Currency)
	fmt.Printf("Duration: %v\n", usage.Duration)
	fmt.Printf("Timestamp: %v\n", usage.Timestamp)
	fmt.Printf("Provider: %s\n\n", usage.Provider)
}

func demoSDKIntegration(tracker *tokentracker.DefaultTokenTracker, openaiProvider, claudeProvider, geminiProvider tokentracker.Provider) {
	fmt.Println("=== SDK Integration ===")

	// Check for environment variables to determine which SDK integrations to demo
	openaiKey := os.Getenv("OPENAI_API_KEY")
	claudeKey := os.Getenv("ANTHROPIC_API_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if openaiKey != "" {
		fmt.Println("Registering OpenAI SDK wrapper...")
		openaiWrapper := sdkwrappers.NewOpenAISDKWrapper(openaiKey, openaiProvider)

		err := tracker.RegisterSDKClient(openaiWrapper)
		if err != nil {
			fmt.Printf("Error registering OpenAI SDK client: %v\n", err)
		} else {
			fmt.Println("OpenAI SDK client registered successfully")
		}
	} else {
		fmt.Println("Skipping OpenAI SDK integration (OPENAI_API_KEY not set)")
	}

	if claudeKey != "" {
		fmt.Println("Registering Claude SDK wrapper...")
		claudeWrapper := sdkwrappers.NewAnthropicSDKWrapper(claudeKey, claudeProvider)

		err := tracker.RegisterSDKClient(claudeWrapper)
		if err != nil {
			fmt.Printf("Error registering Claude SDK client: %v\n", err)
		} else {
			fmt.Println("Claude SDK client registered successfully")
		}
	} else {
		fmt.Println("Skipping Claude SDK integration (ANTHROPIC_API_KEY not set)")
	}

	if geminiKey != "" {
		fmt.Println("Registering Gemini SDK wrapper...")
		geminiWrapper := sdkwrappers.NewGeminiSDKWrapper(geminiKey, geminiProvider)

		err := tracker.RegisterSDKClient(geminiWrapper)
		if err != nil {
			fmt.Printf("Error registering Gemini SDK client: %v\n", err)
		} else {
			fmt.Println("Gemini SDK client registered successfully")
		}
	} else {
		fmt.Println("Skipping Gemini SDK integration (GEMINI_API_KEY not set)")
	}

	// Attempt to update pricing information for all providers
	fmt.Println("Updating pricing information for all providers...")
	if err := tracker.UpdateAllPricing(); err != nil {
		fmt.Printf("Error updating pricing information: %v\n", err)
	} else {
		fmt.Println("Pricing information updated successfully")
	}

	fmt.Println()
}
