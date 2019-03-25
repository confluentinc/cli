package log

import (
	"fmt"
	"io"
	"os"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/hashicorp/go-hclog"
)

// Logger is the standard logger for the Confluent SDK.
type Logger struct {
	params *Params
	l      hclog.Logger
}

var _ ccloud.Logger = (*Logger)(nil)

type Level int

const (
	// For information about unrecoverable events.
	ERROR Level = iota

	// For information about rare but handled events.
	WARN

	// For information about steady state operations.
	INFO

	// For programmer lowlevel analysis.
	DEBUG

	// The most verbose level. Intended to be used for the tracing of actions
	// in code, such as function enters/exits, etc.
	TRACE
)

type Params struct {
	Level  Level
	Output io.Writer
	JSON   bool
}

// New creates a new Logger with the default configuration.
func New() *Logger {
	return NewWithParams(&Params{
		Level:  WARN,
		Output: os.Stderr,
		JSON:   false,
	})
}

// NewWithParams creates and configures a new Logger.
func NewWithParams(params *Params) *Logger {
	return &Logger{
		params: params,
		l: hclog.New(&hclog.LoggerOptions{
			Output:     params.Output,
			JSONFormat: params.JSON,
			Level:      parseLevel(params.Level),
		}),
	}
}

func (l *Logger) Trace(args ...interface{}) {
	if l.l.IsTrace() {
		l.l.Trace(fmt.Sprint(args...))
	}
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	if l.l.IsTrace() {
		l.l.Trace(fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Debug(args ...interface{}) {
	if l.l.IsDebug() {
		l.l.Debug(fmt.Sprint(args...))
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.l.IsDebug() {
		l.l.Debug(fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Info(args ...interface{}) {
	if l.l.IsInfo() {
		l.l.Info(fmt.Sprint(args...))
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	if l.l.IsInfo() {
		l.l.Info(fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Warn(args ...interface{}) {
	if l.l.IsWarn() {
		l.l.Warn(fmt.Sprint(args...))
	}
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.l.IsWarn() {
		l.l.Warn(fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Error(args ...interface{}) {
	if l.l.IsError() {
		l.l.Error(fmt.Sprint(args...))
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.l.IsError() {
		l.l.Error(fmt.Sprintf(format, args...))
	}
}

// Log logs a "msg" and key-value pairs.
// Example: Log("msg", "hello", "key1", "val1", "key2", "val2")
func (l *Logger) Log(args ...interface{}) {
	if l.l.IsDebug() {
		if args[0] != "msg" {
			l.l.Debug(`unexpected logging call, first key should be "msg": ` + fmt.Sprint(args...))
		}
		l.l.Debug(fmt.Sprint(args[1]), args[2:]...)
	}
}

func (l *Logger) SetLevel(level Level) {
	l.params.Level = level
	l.l.SetLevel(parseLevel(level))
}

func parseLevel(level Level) hclog.Level {
	switch level {
	case ERROR:
		return hclog.Error
	case WARN:
		return hclog.Warn
	case INFO:
		return hclog.Info
	case DEBUG:
		return hclog.Debug
	case TRACE:
		return hclog.Trace
	}
	return hclog.NoLevel
}
