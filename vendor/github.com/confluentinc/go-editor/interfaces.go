package editor

// Schema is an interface for validating data.
type Schema interface {
	ValidateBytes(data []byte) error
}

// ValidationFailedFn is a function with which you can handle a validation error.
type ValidationFailedFn func(error) error

// CancelEditingFn is a function with which you can cancel editing and provide a suitable error message.
type CancelEditingFn func() (bool, error)

// PreserveFileFn is a function with which you can inspect the preserved file, edited data, and resulting error.
type PreserveFileFn func(data []byte, file string, err error) ([]byte, string, error)
