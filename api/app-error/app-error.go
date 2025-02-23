package apperror

import "fmt"

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
