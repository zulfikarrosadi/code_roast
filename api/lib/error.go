package lib

type CustomError interface {
	Error() string
}

type AuthError struct {
	Msg  string
	Code int
}

func (authError AuthError) Error() string {
	return authError.Msg
}
