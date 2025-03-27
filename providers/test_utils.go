package providers

// Helper functions shared across provider tests

// StringPtr creates a string pointer from a string value
func StringPtr(s string) *string {
	return &s
}
