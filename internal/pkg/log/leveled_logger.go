package log

// LeveledLogger is a convenience struct for interfacing with the retryable HTTP client used with the v2 Confluent Cloud
// SDK. This is needed because the function names don't line up exactly between libraries.
type LeveledLogger struct {
	unsafeTrace bool
}

func NewLeveledLogger(unsafeTrace bool) *LeveledLogger {
	return &LeveledLogger{unsafeTrace}
}

func (l LeveledLogger) Error(msg string, args ...any) {
	if l.unsafeTrace {
		CliLogger.Errorf(msg, args...)
	}
}

func (l LeveledLogger) Info(msg string, args ...any) {
	if l.unsafeTrace {
		CliLogger.Infof(msg, args...)
	}
}

func (l LeveledLogger) Debug(msg string, args ...any) {
	if l.unsafeTrace {
		CliLogger.Debugf(msg, args...)
	}
}

func (l LeveledLogger) Warn(msg string, args ...any) {
	if l.unsafeTrace {
		CliLogger.Warnf(msg, args...)
	}
}
