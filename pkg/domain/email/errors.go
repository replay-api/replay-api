package email

import "fmt"

type EmailRequiredError struct {
	msg string
}

func NewEmailRequiredError() *EmailRequiredError {
	return &EmailRequiredError{
		msg: "email is required",
	}
}

func (e *EmailRequiredError) Error() string {
	return e.msg
}

type PasswordRequiredError struct {
	msg string
}

func NewPasswordRequiredError() *PasswordRequiredError {
	return &PasswordRequiredError{
		msg: "password is required",
	}
}

func (e *PasswordRequiredError) Error() string {
	return e.msg
}

type VHashRequiredError struct {
	msg string
}

func NewVHashRequiredError() *VHashRequiredError {
	return &VHashRequiredError{
		msg: "vHash is required",
	}
}

func (e *VHashRequiredError) Error() string {
	return e.msg
}

type InvalidVHashError struct {
	msg string
}

func NewInvalidVHashError(vHash string) *InvalidVHashError {
	return &InvalidVHashError{
		msg: fmt.Sprintf("invalid vHash: %s", vHash),
	}
}

func (e *InvalidVHashError) Error() string {
	return e.msg
}

type EmailAlreadyExistsError struct {
	msg string
}

func NewEmailAlreadyExistsError(email string) *EmailAlreadyExistsError {
	return &EmailAlreadyExistsError{
		msg: fmt.Sprintf("email already exists: %s", email),
	}
}

func (e *EmailAlreadyExistsError) Error() string {
	return e.msg
}

type EmailUserNotFoundError struct {
	msg string
}

func NewEmailUserNotFoundError(email string) *EmailUserNotFoundError {
	return &EmailUserNotFoundError{
		msg: fmt.Sprintf("email user not found: %s", email),
	}
}

func (e *EmailUserNotFoundError) Error() string {
	return e.msg
}

type InvalidPasswordError struct {
	msg string
}

func NewInvalidPasswordError() *InvalidPasswordError {
	return &InvalidPasswordError{
		msg: "invalid password",
	}
}

func (e *InvalidPasswordError) Error() string {
	return e.msg
}

type PasswordTooWeakError struct {
	msg string
}

func NewPasswordTooWeakError() *PasswordTooWeakError {
	return &PasswordTooWeakError{
		msg: "password is too weak: minimum 8 characters required",
	}
}

func (e *PasswordTooWeakError) Error() string {
	return e.msg
}

type EmailUserCreationError struct {
	msg string
}

func NewEmailUserCreationError(msg string) *EmailUserCreationError {
	return &EmailUserCreationError{
		msg: fmt.Sprintf("failed to create email user: %s", msg),
	}
}

func (e *EmailUserCreationError) Error() string {
	return e.msg
}
