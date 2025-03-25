package tokentracker

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Cache for token counting to improve performance
type tokenCache struct {
	cache map[string]int
	mu    sync.RWMutex
}

// Global token cache
var globalTokenCache = &tokenCache{
	cache: make(map[string]int),
}

// GetCachedTokenCount gets a cached token count if available
func GetCachedTokenCount(provider, model, text string) (int, bool) {
	globalTokenCache.mu.RLock()
	defer globalTokenCache.mu.RUnlock()

	key := fmt.Sprintf("%s:%s:%s", provider, model, hashString(text))
	count, exists := globalTokenCache.cache[key]
	return count, exists
}

// SetCachedTokenCount sets a token count in the cache
func SetCachedTokenCount(provider, model, text string, count int) {
	globalTokenCache.mu.Lock()
	defer globalTokenCache.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s", provider, model, hashString(text))
	globalTokenCache.cache[key] = count
}

// hashString creates a simple hash of a string for cache keys
// This is a simple implementation and could be improved for production use
func hashString(s string) string {
	if len(s) > 100 {
		// For long strings, just use a prefix and suffix with length
		return fmt.Sprintf("%s...%s:%d", s[:50], s[len(s)-50:], len(s))
	}
	return s
}

// ExtractTextFromMessages extracts all text content from messages
func ExtractTextFromMessages(messages []Message) string {
	var builder strings.Builder

	for _, message := range messages {
		switch content := message.Content.(type) {
		case string:
			builder.WriteString(content)
			builder.WriteString("\n")
		case []ContentPart:
			for _, part := range content {
				if part.Type == "text" {
					builder.WriteString(part.Text)
					builder.WriteString("\n")
				}
			}
		case []interface{}:
			// Handle array of content parts from JSON
			for _, partInterface := range content {
				if part, ok := partInterface.(map[string]interface{}); ok {
					if partType, ok := part["type"].(string); ok && partType == "text" {
						if text, ok := part["text"].(string); ok {
							builder.WriteString(text)
							builder.WriteString("\n")
						}
					}
				}
			}
		}
	}

	return builder.String()
}

// FormatToolsAsJSON formats tools as JSON for token counting
func FormatToolsAsJSON(tools []Tool) string {
	if len(tools) == 0 {
		return ""
	}

	data, err := json.Marshal(tools)
	if err != nil {
		return ""
	}

	return string(data)
}

// EstimateResponseTokens provides a simple estimation of response tokens based on input tokens
// This is a very basic implementation and should be replaced with provider-specific logic
func EstimateResponseTokens(model string, inputTokens int) int {
	// Different models have different response patterns
	// These are very rough estimates and should be refined based on actual usage patterns
	if strings.Contains(model, "gpt-4") {
		return inputTokens // GPT-4 tends to be more verbose
	} else if strings.Contains(model, "gpt-3.5") {
		return inputTokens / 2 // GPT-3.5 is typically less verbose than GPT-4
	} else if strings.Contains(model, "claude") {
		if strings.Contains(model, "opus") {
			return inputTokens * 2 // Claude Opus can be quite verbose
		} else if strings.Contains(model, "sonnet") {
			return inputTokens
		} else {
			return inputTokens / 2 // Claude Haiku is more concise
		}
	} else if strings.Contains(model, "gemini") {
		if strings.Contains(model, "ultra") {
			return inputTokens * 3 / 2 // Gemini Ultra can be verbose
		} else {
			return inputTokens / 2 // Gemini Pro is more concise
		}
	}

	// Default fallback
	return inputTokens / 2
}

// CleanupCache cleans up the token cache to prevent memory leaks
func CleanupCache(maxSize int) {
	globalTokenCache.mu.Lock()
	defer globalTokenCache.mu.Unlock()

	// If cache is smaller than maxSize, do nothing
	if len(globalTokenCache.cache) <= maxSize {
		return
	}

	// Simple strategy: just clear the cache completely
	// A more sophisticated approach would be to use an LRU cache
	globalTokenCache.cache = make(map[string]int)
}