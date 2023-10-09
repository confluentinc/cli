package utils

func WithPanicRecovery(fn func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				OutputErr("Error: internal error occurred")
			}
		}()

		fn()
	}
}
