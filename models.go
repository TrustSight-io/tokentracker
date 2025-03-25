// Package tokentracker provides functionality for tracking token usage and calculating pricing
// for API calls to various LLM providers (Gemini, Claude, OpenAI).
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