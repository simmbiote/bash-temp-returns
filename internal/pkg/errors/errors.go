package errors

import "fmt"

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

const (
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeDatabaseError    = "DATABASE_ERROR"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
)

func NewNotFoundError(resource string, id string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s with ID %s not found", resource, id),
		Err:     nil,
	}
}

func NewInvalidInputError(message string, err error) *AppError {
	return &AppError{Code: ErrCodeInvalidInput, Message: message, Err: err}
}

func NewDatabaseError(message string, err error) *AppError {
	return &AppError{Code: ErrCodeDatabaseError, Message: message, Err: err}
}

func NewValidationError(message string, err error) *AppError {
	return &AppError{Code: ErrCodeValidationFailed, Message: message, Err: err}
}
