package log

import (
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/confluentinc/ccloud-sdk-go"
)

// Logger is the standard logger for the Confluent SDK.
type Logger struct {
	params *Params
	l      hclog.Logger
	buffer []bufferedLog
}

type bufferedLog struct {
	level   Level
	message string
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
	return newLogger(params, hclog.New(&hclog.LoggerOptions{
		Output:     params.Output,
		JSONFormat: params.JSON,
		Level:      parseLevel(params.Level),
	}))
}

func newLogger(params *Params, logger hclog.Logger) *Logger {
	return &Logger{
		params: params,
		l:      logger,
	}
}

func (l *Logger) Named(name string) *Logger {
	logger := l.l.Named(name)
	return newLogger(l.params, logger)
}

func (l *Logger) Trace(args ...interface{}) {
	message := fmt.Sprint(args...)
	if l.l.IsTrace() {
		l.l.Trace(message)
	} else {
		l.bufferLogMessage(TRACE, message)
	}
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.l.IsTrace() {
		l.l.Trace(message)
	} else {
		l.bufferLogMessage(TRACE, message)
	}
}

func (l *Logger) Debug(args ...interface{}) {
	message := fmt.Sprint(args...)
	if l.l.IsDebug() {
		l.l.Debug(message)
	} else {
		l.bufferLogMessage(DEBUG, message)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.l.IsDebug() {
		l.l.Debug(message)
	} else {
		l.bufferLogMessage(DEBUG, message)
	}
}

func (l *Logger) Info(args ...interface{}) {
	message := fmt.Sprint(args...)
	if l.l.IsInfo() {
		l.l.Info(message)
	} else {
		l.bufferLogMessage(INFO, message)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.l.IsInfo() {
		l.l.Info(message)
	} else {
		l.bufferLogMessage(INFO, message)
	}
}

func (l *Logger) Warn(args ...interface{}) {
	message := fmt.Sprint(args...)
	if l.l.IsWarn() {
		l.l.Warn(message)
	} else {
		l.bufferLogMessage(WARN, message)
	}
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.l.IsWarn() {
		l.l.Warn(message)
	} else {
		l.bufferLogMessage(WARN, message)
	}
}

func (l *Logger) Error(args ...interface{}) {
	message := fmt.Sprint(args...)
	if l.l.IsError() {
		l.l.Error(message)
	} else {
		l.bufferLogMessage(ERROR, message)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.l.IsError() {
		l.l.Error(message)
	} else {
		l.bufferLogMessage(ERROR, message)
	}
}

func (l *Logger) bufferLogMessage(level Level, message string) {
	l.buffer = append(l.buffer, bufferedLog{
		level:   level,
		message: message,
	})
}

func (l *Logger) Flush() {
	for _, buffered := range l.buffer {
		if buffered.level < l.GetLevel() {
			continue
		}
		switch buffered.level {
		case ERROR:
			l.Error(buffered.message)
		case WARN:
			l.Warn(buffered.message)
		case INFO:
			l.Info(buffered.message)
		case DEBUG:
			l.Debug(buffered.message)
		case TRACE:
			l.Trace(buffered.message)
		}
	}
	l.buffer = []bufferedLog{}
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

func (l *Logger) GetLevel() Level {
	return l.params.Level
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
