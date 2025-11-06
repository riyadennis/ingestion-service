package foundation

const (
	// ValidationFailed is the error code for
	// validation failures in a POST request
	ValidationFailed = "validation-failure"

	// InvalidRequest is returned if request is not a valid one
	InvalidRequest = "invalid-request"
)

// CustomError holds error code and details about the error
type CustomError struct {
	Code string
	Err  error
}

// Error returns just the error message for a custom error
func (e *CustomError) Error() string {
	return e.Err.Error()
}
