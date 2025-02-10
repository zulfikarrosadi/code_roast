package lib

type CustomError interface {
	Error() string
}

type ServerError struct {
	Msg  string
	Code int
}

func (serverError ServerError) Error() string {
	return serverError.Msg
}
