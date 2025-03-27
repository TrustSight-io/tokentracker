package tokentracker

import (
	"errors"
	"strings"
	"testing"
)

func TestTokenTrackerError_Error(t *testing.T) {
	tests := []struct {
		name        string
		errType     string
		message     string
		cause       error
		expectedMsg string
	}{
		{
			name:        "Error without cause",
			errType:     ErrInvalidModel,
			message:     "model not found",
			cause:       nil,
			expectedMsg: "invalid_model: model not found",
		},
		{
			name:        "Error with cause",
			errType:     ErrTokenizationFailed,
			message:     "failed to tokenize text",
			cause:       errors.New("external error"),
			expectedMsg: "tokenization_failed: failed to tokenize text (cause: external error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.errType, tt.message, tt.cause)

			if err == nil {
				t.Fatal("NewError() returned nil")
			}

			errMsg := err.Error()
			if errMsg != tt.expectedMsg {
				t.Errorf("Error() = %q, expected %q", errMsg, tt.expectedMsg)
			}

			// Test Type field
			if err.Type != tt.errType {
				t.Errorf("err.Type = %q, expected %q", err.Type, tt.errType)
			}

			// Test Message field
			if err.Message != tt.message {
				t.Errorf("err.Message = %q, expected %q", err.Message, tt.message)
			}

			// Test Cause field
			if (err.Cause == nil && tt.cause != nil) || (err.Cause != nil && tt.cause == nil) {
				t.Errorf("err.Cause = %v, expected %v", err.Cause, tt.cause)
			}
			if err.Cause != nil && tt.cause != nil && err.Cause.Error() != tt.cause.Error() {
				t.Errorf("err.Cause = %v, expected %v", err.Cause, tt.cause)
			}
		})
	}
}

func TestTokenTrackerError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := NewError(ErrProviderNotFound, "provider not available", innerErr)

	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, expected %v", unwrapped, innerErr)
	}

	// Test nil cause
	err = NewError(ErrInvalidParams, "invalid parameters", nil)
	unwrapped = err.Unwrap()
	if unwrapped != nil {
		t.Errorf("Unwrap() = %v, expected nil", unwrapped)
	}
}

func TestErrorConstants(t *testing.T) {
	// Check that all error constants are defined
	errorTypes := map[string]string{
		"ErrInvalidModel":       ErrInvalidModel,
		"ErrInvalidParams":      ErrInvalidParams,
		"ErrProviderNotFound":   ErrProviderNotFound,
		"ErrTokenizationFailed": ErrTokenizationFailed,
		"ErrPricingNotFound":    ErrPricingNotFound,
	}

	for name, errType := range errorTypes {
		if errType == "" {
			t.Errorf("Expected %s to be defined", name)
		}

		// Check that error type follows the naming convention
		if !strings.Contains(errType, "_") {
			t.Errorf("Expected %s to use snake_case format, got: %s", name, errType)
		}
	}
}

func TestNewError(t *testing.T) {
	// Test creating different error types
	errorTypes := []string{
		ErrInvalidModel,
		ErrInvalidParams,
		ErrProviderNotFound,
		ErrTokenizationFailed,
		ErrPricingNotFound,
	}

	for _, errType := range errorTypes {
		t.Run(errType, func(t *testing.T) {
			message := "test message"
			cause := errors.New("test cause")

			err := NewError(errType, message, cause)

			if err.Type != errType {
				t.Errorf("NewError().Type = %q, expected %q", err.Type, errType)
			}
			if err.Message != message {
				t.Errorf("NewError().Message = %q, expected %q", err.Message, message)
			}
			if err.Cause != cause {
				t.Errorf("NewError().Cause = %v, expected %v", err.Cause, cause)
			}
		})
	}
}

func TestErrors_Integration(t *testing.T) {
	// Create a chain of errors
	innerErr := errors.New("database connection failed")
	middleErr := NewError(ErrTokenizationFailed, "could not tokenize text", innerErr)
	outerErr := NewError(ErrInvalidParams, "invalid request", middleErr)

	// Check the error message
	expected := "invalid_params: invalid request (cause: tokenization_failed: could not tokenize text (cause: database connection failed))"
	if outerErr.Error() != expected {
		t.Errorf("Error() = %q, expected %q", outerErr.Error(), expected)
	}

	// Check that the cause is correctly set
	if outerErr.Cause != middleErr {
		t.Errorf("outerErr.Cause = %v, expected %v", outerErr.Cause, middleErr)
	}

	// Check that errors.Is works correctly
	if !errors.Is(outerErr, middleErr) {
		t.Errorf("errors.Is(outerErr, middleErr) = false, expected true")
	}
}
