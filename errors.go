package siocore

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	DecryptionFailed      = "Decryption failed - unable to proceed error: %v"
	InvalidEmailErr       = "invalid email"
	InvalidPasswordErr    = "invalid password: Requirements are 8 char min, 1 upper, 1 special, and 1 numerical"
	InvalidNameErr        = "invalid name"
	InvalidDescriptionErr = "invalid description"
	InvalidPhoneErr       = "please enter a ten digit mobile number"
	InvalidDobErr         = "please enter a valid date for dob"
	InvalidAgeErr         = "customer must be at least 13 years of age"
	InvalidEndDateErr     = "end date must be after start date"
	TokenInvalid          = "unauthorized"
	INVALID_ID            = "id must be numerical"
)

type AppError struct {
	error
	Code int
}

func NewAppError(message string, code int) *AppError {
	return &AppError{
		error: fmt.Errorf(message),
		Code:  code,
	}
}

func (ae *AppError) FromString(err string, code int) {
	ae.error = fmt.Errorf(err)
	ae.Code = code
}

func (ae *AppError) FromError(err error, code int) {
	ae.error = err
	ae.Code = code
}

func (e *AppError) Error() string {
	return e.error.Error()
}

func NewNotFoundError(message string) *AppError {
	return NewAppError(message, 404)
}

func NewBadRequestError(message string) *AppError {
	return NewAppError(message, 400)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(message, 401)
}

func NewInternalServerError(message string) *AppError {
	return NewAppError(message, 500)
}

// GetRuntimeStack Returns the stack trace of the current goroutine
func GetRuntimeStack() string {
	stackBuf := make([]byte, 1024)
	stackSize := runtime.Stack(stackBuf, false)
	stackBuf = stackBuf[:stackSize]
	return strings.TrimSpace(string(stackBuf))
}
