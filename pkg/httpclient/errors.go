package httpclient

import "fmt"

// RetryError is returned when a request fails without a response and wrapping adds information.
type RetryError struct {
	Attempts     int
	Err          error
	MaybeApplied bool // a POST or PATCH that may have reached the server
}

func (e *RetryError) Error() string {
	switch {
	case e.MaybeApplied && e.Attempts == 1:
		return fmt.Sprintf("request failed but may have reached the server: %v", e.Err)
	case e.MaybeApplied:
		return fmt.Sprintf("request failed after %d attempts but may have reached the server: %v", e.Attempts, e.Err)
	default:
		return fmt.Sprintf("request failed after %d attempts: %v", e.Attempts, e.Err)
	}
}

func (e *RetryError) Unwrap() error {
	return e.Err
}

func (e *RetryError) GetSuggestionsMsg() string {
	if e.MaybeApplied {
		return "If this command makes a change, check whether it was applied, for example with the matching `list` or `describe` command, before trying again."
	}
	return "Check your network connection and proxy settings, then try again."
}
