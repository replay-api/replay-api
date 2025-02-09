package common

import "fmt"

func NewErrUnauthorized() error {
	return fmt.Errorf("Unauthorized")
}
