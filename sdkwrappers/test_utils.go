package sdkwrappers

// Helper functions shared across SDK wrapper tests

// StringPtr creates a string pointer from a string value
func StringPtr(s string) *string {
	return &s
}
