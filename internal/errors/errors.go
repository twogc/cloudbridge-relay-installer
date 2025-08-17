package errors

import (
	"fmt"
	"net"
	"strings"
)

// ErrorType represents the type of error.
type ErrorType string

const (
	// Connection errors.
	ErrConnectionFailed  ErrorType = "CONNECTION_FAILED"
	ErrConnectionClosed  ErrorType = "CONNECTION_CLOSED"
	ErrConnectionTimeout ErrorType = "CONNECTION_TIMEOUT"

	// Protocol errors.
	ErrInvalidMessage ErrorType = "INVALID_MESSAGE"
	ErrInvalidToken   ErrorType = "INVALID_TOKEN"
	ErrRateLimit      ErrorType = "RATE_LIMIT"

	// System errors
	ErrSystemError ErrorType = "SYSTEM_ERROR"
)

// RelayError represents a relay error
type RelayError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface
func (e *RelayError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *RelayError) Unwrap() error {
	return e.Err
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for net.Error
	if _, ok := err.(net.Error); ok {
		return true
	}

	// Check for our custom errors
	if relayErr, ok := err.(*RelayError); ok {
		return strings.HasPrefix(string(relayErr.Type), "CONNECTION_")
	}

	return false
}

// IsProtocolError checks if the error is a protocol error
func IsProtocolError(err error) bool {
	if err == nil {
		return false
	}

	if relayErr, ok := err.(*RelayError); ok {
		return strings.HasPrefix(string(relayErr.Type), "INVALID_")
	}

	return false
}

// IsSystemError checks if the error is a system error
func IsSystemError(err error) bool {
	if err == nil {
		return false
	}

	if relayErr, ok := err.(*RelayError); ok {
		return strings.HasPrefix(string(relayErr.Type), "SYSTEM_")
	}

	return false
}

// NewConnectionError creates a new connection error
func NewConnectionError(err error) *RelayError {
	return &RelayError{
		Type:    ErrConnectionFailed,
		Message: "Failed to establish connection",
		Err:     err,
	}
}

// NewProtocolError creates a new protocol error
func NewProtocolError(message string, err error) *RelayError {
	return &RelayError{
		Type:    ErrInvalidMessage,
		Message: message,
		Err:     err,
	}
}

// NewTokenError creates a new token error
func NewTokenError(err error) *RelayError {
	return &RelayError{
		Type:    ErrInvalidToken,
		Message: "Invalid token",
		Err:     err,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError() *RelayError {
	return &RelayError{
		Type:    ErrRateLimit,
		Message: "Rate limit exceeded",
	}
}

// NewSystemError creates a new system error
func NewSystemError(message string, err error) *RelayError {
	return &RelayError{
		Type:    ErrSystemError,
		Message: message,
		Err:     err,
	}
}
