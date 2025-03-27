package tokentracker

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestTokenCache(t *testing.T) {
	// Clear the cache before testing
	globalTokenCache.mu.Lock()
	globalTokenCache.cache = make(map[string]int)
	globalTokenCache.mu.Unlock()

	// Test GetCachedTokenCount with empty cache
	_, exists := GetCachedTokenCount("test-provider", "test-model", "test text")
	if exists {
		t.Error("Expected GetCachedTokenCount to return false for empty cache")
	}

	// Test SetCachedTokenCount
	SetCachedTokenCount("test-provider", "test-model", "test text", 10)

	// Test GetCachedTokenCount with populated cache
	count, exists := GetCachedTokenCount("test-provider", "test-model", "test text")
	if !exists {
		t.Error("Expected GetCachedTokenCount to return true after setting cache")
	}
	if count != 10 {
		t.Errorf("Expected cached count to be 10, got %d", count)
	}

	// Test cache with different providers/models but same text
	SetCachedTokenCount("other-provider", "test-model", "test text", 20)
	SetCachedTokenCount("test-provider", "other-model", "test text", 30)

	count, _ = GetCachedTokenCount("other-provider", "test-model", "test text")
	if count != 20 {
		t.Errorf("Expected cached count to be 20 for other-provider, got %d", count)
	}

	count, _ = GetCachedTokenCount("test-provider", "other-model", "test text")
	if count != 30 {
		t.Errorf("Expected cached count to be 30 for other-model, got %d", count)
	}

	// Test with empty provider and model
	SetCachedTokenCount("", "", "test text", 40)
	count, exists = GetCachedTokenCount("", "", "test text")
	if !exists || count != 40 {
		t.Errorf("Expected cached count to be 40 for empty provider/model, got exists=%v, count=%d", exists, count)
	}
}

func TestHashString(t *testing.T) {
	// Test with short string
	shortStr := "short string"
	shortHash := hashString(shortStr)
	if shortHash != shortStr {
		t.Errorf("Expected hashString to return the same string for short strings, got %q", shortHash)
	}

	// Test with long string
	longStr := strings.Repeat("a", 200)
	longHash := hashString(longStr)
	if longHash == longStr {
		t.Errorf("Expected hashString to modify long strings")
	}

	// Check that the hash includes the length
	if !strings.Contains(longHash, "200") {
		t.Errorf("Expected hash of long string to include the string length")
	}

	// Check that the hash includes both prefix and suffix
	if !strings.Contains(longHash, "a...a") {
		t.Errorf("Expected hash of long string to include prefix and suffix")
	}

	// Test with exactly 100 characters
	str100 := strings.Repeat("b", 100)
	hash100 := hashString(str100)
	if hash100 != str100 {
		t.Errorf("Expected hashString to return the same string for strings of exactly 100 chars")
	}

	// Test with 101 characters
	str101 := strings.Repeat("c", 101)
	hash101 := hashString(str101)
	if hash101 == str101 {
		t.Errorf("Expected hashString to modify strings longer than 100 chars")
	}
	if !strings.Contains(hash101, "101") {
		t.Errorf("Expected hash of 101-char string to include the string length")
	}
}

func TestExtractTextFromMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		expected string
	}{
		{
			name:     "Empty messages",
			messages: []Message{},
			expected: "",
		},
		{
			name: "Simple string content",
			messages: []Message{
				{
					Role:    "user",
					Content: "Hello",
				},
				{
					Role:    "assistant",
					Content: "World",
				},
			},
			expected: "Hello\nWorld\n",
		},
		{
			name: "Content parts",
			messages: []Message{
				{
					Role: "user",
					Content: []ContentPart{
						{
							Type: "text",
							Text: "Hello",
						},
						{
							Type: "text",
							Text: "World",
						},
					},
				},
			},
			expected: "Hello\nWorld\n",
		},
		{
			name: "Mixed content types",
			messages: []Message{
				{
					Role:    "system",
					Content: "System message",
				},
				{
					Role: "user",
					Content: []ContentPart{
						{
							Type: "text",
							Text: "User message",
						},
						{
							Type: "image",
							Text: "", // Should be ignored
						},
					},
				},
			},
			expected: "System message\nUser message\n",
		},
		{
			name: "JSON array content",
			messages: []Message{
				{
					Role: "user",
					Content: []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": "JSON content",
						},
						map[string]interface{}{
							"type": "image",
							"url":  "http://example.com/image.jpg",
						},
					},
				},
			},
			expected: "JSON content\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTextFromMessages(tt.messages)
			if result != tt.expected {
				t.Errorf("ExtractTextFromMessages() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestFormatToolsAsJSON(t *testing.T) {
	tests := []struct {
		name     string
		tools    []Tool
		expected string
		isEmpty  bool
	}{
		{
			name:     "Empty tools",
			tools:    []Tool{},
			expected: "",
			isEmpty:  true,
		},
		{
			name: "Simple tools",
			tools: []Tool{
				{
					Type: "function",
					Function: map[string]interface{}{
						"name":        "get_weather",
						"description": "Get the weather",
					},
				},
			},
			isEmpty: false,
		},
		{
			name: "Multiple tools",
			tools: []Tool{
				{
					Type: "function",
					Function: map[string]interface{}{
						"name":        "get_weather",
						"description": "Get the weather",
					},
				},
				{
					Type: "function",
					Function: map[string]interface{}{
						"name":        "get_location",
						"description": "Get the location",
					},
				},
			},
			isEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToolsAsJSON(tt.tools)
			if tt.isEmpty {
				if result != "" {
					t.Errorf("FormatToolsAsJSON() = %q, expected empty string", result)
				}
			} else {
				// Verify the result is valid JSON
				var parsed interface{}
				err := json.Unmarshal([]byte(result), &parsed)
				if err != nil {
					t.Errorf("FormatToolsAsJSON() produced invalid JSON: %v", err)
				}

				// Check that tools are included in the JSON
				jsonString := string(result)
				if !strings.Contains(jsonString, "function") {
					t.Errorf("FormatToolsAsJSON() result doesn't contain expected content: %s", jsonString)
				}
			}
		})
	}
}

func TestEstimateResponseTokens(t *testing.T) {
	tests := []struct {
		name        string
		model       string
		inputTokens int
		minExpected int
		maxExpected int
	}{
		{
			name:        "GPT-4",
			model:       "gpt-4",
			inputTokens: 100,
			minExpected: 100,
			maxExpected: 100,
		},
		{
			name:        "GPT-3.5",
			model:       "gpt-3.5-turbo",
			inputTokens: 100,
			minExpected: 50,
			maxExpected: 50,
		},
		{
			name:        "Claude Opus",
			model:       "claude-3-opus",
			inputTokens: 100,
			minExpected: 200,
			maxExpected: 200,
		},
		{
			name:        "Claude Sonnet",
			model:       "claude-3-sonnet",
			inputTokens: 100,
			minExpected: 100,
			maxExpected: 100,
		},
		{
			name:        "Claude Haiku",
			model:       "claude-3-haiku",
			inputTokens: 100,
			minExpected: 50,
			maxExpected: 50,
		},
		{
			name:        "Gemini Ultra",
			model:       "gemini-ultra",
			inputTokens: 100,
			minExpected: 150,
			maxExpected: 150,
		},
		{
			name:        "Gemini Pro",
			model:       "gemini-pro",
			inputTokens: 100,
			minExpected: 50,
			maxExpected: 50,
		},
		{
			name:        "Unknown model",
			model:       "unknown-model",
			inputTokens: 100,
			minExpected: 50,
			maxExpected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateResponseTokens(tt.model, tt.inputTokens)
			if result < tt.minExpected || result > tt.maxExpected {
				t.Errorf("EstimateResponseTokens(%q, %d) = %d, expected between %d and %d",
					tt.model, tt.inputTokens, result, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestCleanupCache(t *testing.T) {
	// Populate the cache with some entries
	globalTokenCache.mu.Lock()
	globalTokenCache.cache = make(map[string]int)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key-%d", i)
		globalTokenCache.cache[key] = i
	}
	size := len(globalTokenCache.cache)
	globalTokenCache.mu.Unlock()

	// Verify initial size
	if size != 10 {
		t.Errorf("Expected initial cache size to be 10, got %d", size)
	}

	// Test cleanup with larger max size (should not clean up)
	CleanupCache(20)
	globalTokenCache.mu.RLock()
	size = len(globalTokenCache.cache)
	globalTokenCache.mu.RUnlock()
	if size != 10 {
		t.Errorf("Expected cache size to remain 10 after CleanupCache(20), got %d", size)
	}

	// Test cleanup with smaller max size (should clean up)
	CleanupCache(5)
	globalTokenCache.mu.RLock()
	size = len(globalTokenCache.cache)
	globalTokenCache.mu.RUnlock()
	if size != 0 {
		t.Errorf("Expected cache to be emptied after CleanupCache(5), got size %d", size)
	}
}
