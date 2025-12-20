package common

import (
	"fmt"
)

// Error types for type assertions
type ErrUnauthorized struct {
	message string
}

func (e *ErrUnauthorized) Error() string {
	return e.message
}

type ErrForbidden struct {
	message string
}

func (e *ErrForbidden) Error() string {
	return e.message
}

type ErrNotFound struct {
	message string
}

func (e *ErrNotFound) Error() string {
	return e.message
}

type ErrAlreadyExists struct {
	message string
}

func (e *ErrAlreadyExists) Error() string {
	return e.message
}

type ErrInvalidInput struct {
	message string
}

func (e *ErrInvalidInput) Error() string {
	return e.message
}

func NewErrUnauthorized() error {
	return &ErrUnauthorized{message: "Unauthorized"}
}

func NewErrForbidden(messages ...string) error {
	msg := "Forbidden"
	if len(messages) > 0 && messages[0] != "" {
		msg = messages[0]
	}
	return &ErrForbidden{message: msg}
}

func NewErrAlreadyExists(resourceType ResourceType, fieldName string, value interface{}) error {
	return &ErrAlreadyExists{message: fmt.Sprintf("%s with %s %v already exists", resourceType, fieldName, value)}
}

func NewErrNotFound(resourceType ResourceType, fieldName string, value interface{}) error {
	return &ErrNotFound{message: fmt.Sprintf("%s with %s %v not found", resourceType, fieldName, value)}
}

func NewErrInvalidInput(message string) error {
	return &ErrInvalidInput{message: message}
}

type ErrBadRequest struct {
	message string
}

func (e *ErrBadRequest) Error() string {
	return e.message
}

func NewErrBadRequest(message string) error {
	return &ErrBadRequest{message: message}
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	_, ok := err.(*ErrNotFound)
	return ok
}

// IsUnauthorizedError checks if an error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	_, ok := err.(*ErrUnauthorized)
	return ok
}

// IsForbiddenError checks if an error is a forbidden error
func IsForbiddenError(err error) bool {
	_, ok := err.(*ErrForbidden)
	return ok
}

// IsBadRequestError checks if an error is a bad request error
func IsBadRequestError(err error) bool {
	_, ok := err.(*ErrBadRequest)
	return ok
}

// IsInvalidInputError checks if an error is an invalid input error
func IsInvalidInputError(err error) bool {
	_, ok := err.(*ErrInvalidInput)
	return ok
}
