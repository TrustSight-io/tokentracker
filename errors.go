package tokentracker

import "fmt"

// Error types
const (
	ErrInvalidModel       = "invalid_model"
	ErrInvalidParams      = "invalid_params"
	ErrProviderNotFound   = "provider_not_found"
	ErrTokenizationFailed = "tokenization_failed"
	ErrPricingNotFound    = "pricing_not_found"
)

// TokenTrackerError represents an error in the token tracker
type TokenTrackerError struct {
	Type    string
	Message string
	Cause   error
}

// Error returns the error message
func (e *TokenTrackerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *TokenTrackerError) Unwrap() error {
	return e.Cause
}

// NewError creates a new TokenTrackerError
func NewError(errType, message string, cause error) *TokenTrackerError {
	return &TokenTrackerError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}