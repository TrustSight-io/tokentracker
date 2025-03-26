// Package common contains shared types used across the tokentracker project
package common

import "time"

// ModelPricing contains pricing information for a specific model
type ModelPricing struct {
	InputPricePerToken  float64
	OutputPricePerToken float64
	Currency            string
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

// TokenUsage represents token usage information extracted from API responses
type TokenUsage struct {
	InputTokens    int
	OutputTokens   int
	TotalTokens    int
	CompletionID   string
	Model          string
	Timestamp      time.Time
	PromptTokens   int    // Some APIs use "prompt" instead of "input"
	ResponseTokens int    // Some APIs use "response" instead of "output"
	RequestID      string // Some APIs provide a request ID
}