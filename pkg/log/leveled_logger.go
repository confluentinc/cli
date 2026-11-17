package log

import (
	"fmt"
	"regexp"
)

// LeveledLogger adapts CliLogger to retryablehttp's LeveledLogger, which passes key/value pairs after the message.
// Every level goes to Debug, so a failed attempt that a retry recovers from doesn't print at the default ERROR level.
type LeveledLogger struct{}

func NewLeveledLogger() *LeveledLogger {
	return &LeveledLogger{}
}

func (LeveledLogger) Error(msg string, keysAndValues ...any) { logAtDebug(msg, keysAndValues) }
func (LeveledLogger) Warn(msg string, keysAndValues ...any)  { logAtDebug(msg, keysAndValues) }
func (LeveledLogger) Info(msg string, keysAndValues ...any)  { logAtDebug(msg, keysAndValues) }
func (LeveledLogger) Debug(msg string, keysAndValues ...any) { logAtDebug(msg, keysAndValues) }

var urlQuery = regexp.MustCompile(`(https?://[^\s"?#]*)\?[^\s"#]*`)

func logAtDebug(msg string, keysAndValues []any) {
	if CliLogger.Level < UNSAFE_TRACE {
		// Query strings can carry secrets (signed URLs, tokens), so only UNSAFE_TRACE shows them.
		hidden := make([]any, 0, len(keysAndValues))
		for _, arg := range keysAndValues {
			hidden = append(hidden, urlQuery.ReplaceAllString(fmt.Sprint(arg), "$1"))
		}
		keysAndValues = hidden
	}
	CliLogger.Log(msg, keysAndValues...)
}
