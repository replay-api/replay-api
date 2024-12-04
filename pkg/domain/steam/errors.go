package steam

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

// Steam User Not Found Error
type SteamUserNotFoundError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *SteamUserNotFoundError) Error() string {
	return e.Message
}

// NewSteamUserNotFoundError creates a new SteamUserNotFoundError
func NewSteamUserNotFoundError(message string) *SteamUserNotFoundError {
	return &SteamUserNotFoundError{
		Message: message,
	}
}

// Steam User Already Exists Error
type SteamUserAlreadyExistsError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *SteamUserAlreadyExistsError) Error() string {
	return e.Message
}

// NewSteamUserAlreadyExistsError creates a new SteamUserAlreadyExistsError
func NewSteamUserAlreadyExistsError(message string) *SteamUserAlreadyExistsError {
	return &SteamUserAlreadyExistsError{
		Message: message,
	}
}

// Steam User Creation Error
type SteamUserCreationError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *SteamUserCreationError) Error() string {
	return e.Message
}

// NewSteamUserCreationError creates a new SteamUserCreationError
func NewSteamUserCreationError(message string) *SteamUserCreationError {
	return &SteamUserCreationError{
		Message: message,
	}
}

// Steam User Verification Error
type SteamUserVerificationError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *SteamUserVerificationError) Error() string {
	return e.Message
}

// NewSteamUserVerificationError creates a new SteamUserVerificationError
func NewSteamUserVerificationError(expectedSteamID uint64, receivedSteamID uint64) *SteamUserVerificationError {
	return &SteamUserVerificationError{
		Message: fmt.Sprintf("SteamID verification failed. Expected: %d, Received: %d", expectedSteamID, receivedSteamID),
	}
}

type SteamIDMismatchError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *SteamIDMismatchError) Error() string {
	return e.Message
}

// NewSteamIDMismatchError creates a new SteamIDMismatchError
func NewSteamIDMismatchError(receivedSteamID string) *SteamIDMismatchError {
	return &SteamIDMismatchError{
		Message: fmt.Sprintf("SteamID mismatch: %s", receivedSteamID),
	}
}

type SteamIDRequiredError struct {
	// Error message
	Message string
}

// Error returns the error message
func (e *SteamIDRequiredError) Error() string {
	return e.Message
}

// NewSteamIDRequiredError creates a new SteamIDRequiredError
func NewSteamIDRequiredError() *SteamIDRequiredError {
	return &SteamIDRequiredError{
		Message: "SteamID is required",
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
