package command

type silentError struct {
	message string
}

func (e silentError) Error() string {
	return e.message
}

func newSilentError(message string) error {
	return silentError{message: message}
}

func IsSilentError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(silentError)
	return ok
}
