package base

// Error is an error with a user-facing message.
type Error struct {
	Msg string
}

// Error returns the message unchanged.
func (e *Error) Error() string {
	return e.Msg
}

// Fail creates an Error containing msg.
func Fail(msg string) error {
	return &Error{Msg: msg}
}
