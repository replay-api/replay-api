package common

import (
	"fmt"
)

func NewErrUnauthorized() error {
	return fmt.Errorf("Unauthorized")
}

func NewErrAlreadyExists(resourceType ResourceType, fieldName string, value interface{}) error {
	return fmt.Errorf("%s with %s %v already exists", resourceType, fieldName, value)
}

func NewErrNotFound(resourceType ResourceType, fieldName string, value interface{}) error {
	return fmt.Errorf("%s with %s %v not found", resourceType, fieldName, value)
}
