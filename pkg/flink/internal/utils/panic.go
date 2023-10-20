package utils

func WithPanicRecovery(fn func()) func() {
	return WithCustomPanicRecovery(fn, func() {
		OutputErr("Error: internal error occurred")
	})
}

func WithCustomPanicRecovery(fn func(), customRecovery func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				if customRecovery != nil {
					WithPanicRecovery(customRecovery)()
				}
			}
		}()
		fn()
	}
}
