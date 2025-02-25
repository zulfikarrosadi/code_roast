package apperror

import (
	"fmt"

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

type ErrorDetail struct {
	Path    []string `json:"path"`
	Value   string   `json:"value"`
	Message string   `json:"message"`
}

const (
	VALIDATION_ERROR = "validation error"
)

func ValidateError(validationError validator.ValidationErrors) []ErrorDetail {
	errorDetails := []ErrorDetail{}

	for _, fieldError := range validationError {
		switch fieldError.Tag() {
		case "required":
			errorDetail := ErrorDetail{
				Path:    []string{fieldError.Field()},
				Message: fieldError.Field() + " is required",
			}
			errorDetails = append(errorDetails, errorDetail)
		case "email":
			errorDetail := ErrorDetail{
				Path:    []string{fieldError.Field()},
				Message: "invalid email format",
			}
			errorDetails = append(errorDetails, errorDetail)
		case "eqfield":
			fmt.Println(fieldError.Field(), fieldError.StructField())
			if fieldError.Field() == "passwordConfirmation" {
				errorDetail := ErrorDetail{
					Path:    []string{"password", fieldError.Field()},
					Message: "password and password confirmation is not match",
				}
				errorDetails = append(errorDetails, errorDetail)
			}
		}
	}
	return errorDetails
}
