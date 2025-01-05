package google

import "fmt"

// Invalid VHash Error
type InvalidVHashError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *InvalidVHashError) Error() string {
	return e.Message
}

// NewInvalidVHashError creates a new InvalidVHashError
func NewInvalidVHashError(invalidVHash string) *InvalidVHashError {
	return &InvalidVHashError{
		Message: fmt.Sprintf("Invalid vHash: %s", invalidVHash),
	}
}

// Google User Not Found Error
type GoogleUserNotFoundError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *GoogleUserNotFoundError) Error() string {
	return e.Message
}

// NewGoogleUserNotFoundError creates a new GoogleUserNotFoundError
func NewGoogleUserNotFoundError(message string) *GoogleUserNotFoundError {
	return &GoogleUserNotFoundError{
		Message: message,
	}
}

// Google User Already Exists Error
type GoogleUserAlreadyExistsError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *GoogleUserAlreadyExistsError) Error() string {
	return e.Message
}

// NewGoogleUserAlreadyExistsError creates a new GoogleUserAlreadyExistsError
func NewGoogleUserAlreadyExistsError(message string) *GoogleUserAlreadyExistsError {
	return &GoogleUserAlreadyExistsError{
		Message: message,
	}
}

// Google User Creation Error
type GoogleUserCreationError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *GoogleUserCreationError) Error() string {
	return e.Message
}

// NewGoogleUserCreationError creates a new GoogleUserCreationError
func NewGoogleUserCreationError(message string) *GoogleUserCreationError {
	return &GoogleUserCreationError{
		Message: message,
	}
}

// Google User Verification Error
type GoogleUserVerificationError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *GoogleUserVerificationError) Error() string {
	return e.Message
}

// NewGoogleUserVerificationError creates a new GoogleUserVerificationError
func NewGoogleUserVerificationError(expectedGoogleID uint64, receivedGoogleID uint64) *GoogleUserVerificationError {
	return &GoogleUserVerificationError{
		Message: fmt.Sprintf("GoogleID verification failed. Expected: %d, Received: %d", expectedGoogleID, receivedGoogleID),
	}
}

type GoogleIDMismatchError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *GoogleIDMismatchError) Error() string {
	return e.Message
}

// NewGoogleIDMismatchError creates a new GoogleIDMismatchError
func NewGoogleIDMismatchError(receivedGoogleID string) *GoogleIDMismatchError {
	return &GoogleIDMismatchError{
		Message: fmt.Sprintf("GoogleID mismatch: %s", receivedGoogleID),
	}
}

type GoogleIDRequiredError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *GoogleIDRequiredError) Error() string {
	return e.Message
}

// NewGoogleIDRequiredError creates a new GoogleIDRequiredError
func NewGoogleIDRequiredError() *GoogleIDRequiredError {
	return &GoogleIDRequiredError{
		Message: "GoogleID is required",
	}
}

type VHashRequiredError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *VHashRequiredError) Error() string {
	return e.Message
}

// NewVHashRequiredError creates a new VHashRequiredError
func NewVHashRequiredError() *VHashRequiredError {
	return &VHashRequiredError{
		Message: "vHash is required",
	}
}
