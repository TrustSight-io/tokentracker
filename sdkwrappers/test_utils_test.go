package sdkwrappers

import (
	"testing"
)

func TestStringPtr(t *testing.T) {
	// Test with empty string
	emptyStr := ""
	ptr := StringPtr(emptyStr)
	if ptr == nil {
		t.Fatal("StringPtr returned nil for empty string")
	}
	if *ptr != emptyStr {
		t.Errorf("StringPtr() = %q, want %q", *ptr, emptyStr)
	}

	// Test with non-empty string
	testStr := "test string"
	ptr = StringPtr(testStr)
	if ptr == nil {
		t.Fatal("StringPtr returned nil for non-empty string")
	}
	if *ptr != testStr {
		t.Errorf("StringPtr() = %q, want %q", *ptr, testStr)
	}
}
