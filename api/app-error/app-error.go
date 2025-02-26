package apperror

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func New(code int, message string, err error) error {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

const (
	VALIDATION_ERROR = "validation error"
)

type ErrorDetails map[string]string

func HandlerValidatorError(validationError validator.ValidationErrors) ErrorDetails {
	errorDetails := make(ErrorDetails)
	for _, fieldError := range validationError {
		fmt.Printf("tag: %s, field: %s", fieldError.Tag(), fieldError.Field())
		switch fieldError.Tag() {
		case "required":
			errorDetails[strings.ToLower(fieldError.Field())] = fieldError.Field() + " is required"
		case "email":
			errorDetails[strings.ToLower(fieldError.Field())] = "invalid email format"
		case "eqfield":
			if fieldError.Field() == "PasswordConfirmation" {
				errorDetails["password"] = "password and password confirmation not match"
			}
		}
	}
	return errorDetails
}
