package lib

type CustomError interface {
	Error() string
}

type ServerError struct {
	Msg string
}

func (serverError ServerError) Error() string {
	return serverError.Msg
}
