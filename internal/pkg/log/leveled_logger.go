package log

// LeveledLogger is a convenience struct for interfacing with the retryable HTTP client used with the v2 Confluent Cloud
// SDK. This is needed because the function names don't line up exactly between libraries.
type LeveledLogger struct{}

func (LeveledLogger) Error(msg string, args ...interface{}) {
	CliLogger.Errorf(msg, args...)
}

func (LeveledLogger) Info(msg string, args ...interface{}) {
	CliLogger.Infof(msg, args...)
}

func (LeveledLogger) Debug(msg string, args ...interface{}) {
	CliLogger.Debugf(msg, args...)
}

func (LeveledLogger) Warn(msg string, args ...interface{}) {
	CliLogger.Warnf(msg, args...)
}
