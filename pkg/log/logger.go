package log

import (
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/go-hclog"
)

// TODO: once we migrate from ccloud-sdk-v1 we should change these functions to act on the
// TODO: global logger instead of (l *Logger) and then we can call log.Debug() instead of log.CliLogger.Debug()

func init() {
	CliLogger = New(ERROR, os.Stderr)
}

// CliLogger is a global logger instance
var CliLogger *Logger

// Logger is the standard logger for the Confluent CLI and is a wrapper around go-hclog
type Logger struct {
	Level  Level
	logger hclog.Logger
	buffer []leveledMessage
}

type leveledMessage struct {
	level   Level
	message string
}

type Level int

const (
	// ERROR is for information about unrecoverable events
	ERROR Level = iota
	// WARN is for information about rare but handled events
	WARN
	// INFO is for information about steady state operations
	INFO
	// DEBUG is for programmer low-level analysis
	DEBUG
	// TRACE is intended to be used for the tracing of actions in code, such as function enters/exits, etc
	TRACE
)

// New creates and configures a new Logger
func New(level Level, output io.Writer) *Logger {
	return &Logger{
		Level: level,
		logger: hclog.New(&hclog.LoggerOptions{
			Output: output,
			Level:  mapToHclogLevel(level),
		}),
	}
}

func (l *Logger) SetVerbosity(verbosity int) {
	level := Level(verbosity)
	if verbosity > int(TRACE) {
		level = TRACE
	}

	l.Level = level
	l.logger.SetLevel(mapToHclogLevel(level))
}

func (l *Logger) Trace(args ...any) {
	message := fmt.Sprint(args...)
	if l.logger.IsTrace() {
		l.logger.Trace(message)
	} else {
		l.append(TRACE, message)
	}
}

func (l *Logger) Tracef(format string, args ...any) {
	l.Trace(fmt.Sprintf(format, args...))
}

func (l *Logger) Debug(args ...any) {
	message := fmt.Sprint(args...)
	if l.logger.IsDebug() {
		l.logger.Debug(message)
	} else {
		l.append(DEBUG, message)
	}
}

func (l *Logger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *Logger) Info(args ...any) {
	message := fmt.Sprint(args...)
	if l.logger.IsInfo() {
		l.logger.Info(message)
	} else {
		l.append(INFO, message)
	}
}

func (l *Logger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(args ...any) {
	message := fmt.Sprint(args...)
	if l.logger.IsWarn() {
		l.logger.Warn(message)
	} else {
		l.append(WARN, message)
	}
}

func (l *Logger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Error(args ...any) {
	message := fmt.Sprint(args...)
	if l.logger.IsError() {
		l.logger.Error(message)
	} else {
		l.append(ERROR, message)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) append(level Level, message string) {
	l.buffer = append(l.buffer, leveledMessage{level, message})
}

func (l *Logger) Flush() {
	for _, lm := range l.buffer {
		if lm.level < l.Level {
			continue
		}

		switch lm.level {
		case ERROR:
			l.Error(lm.message)
		case WARN:
			l.Warn(lm.message)
		case INFO:
			l.Info(lm.message)
		case DEBUG:
			l.Debug(lm.message)
		case TRACE:
			l.Trace(lm.message)
		}
	}

	l.buffer = []leveledMessage{}
}

// Log logs a "msg" and key-value pairs.
// Example: Log("msg", "hello", "key1", "val1", "key2", "val2")
func (l *Logger) Log(args ...any) {
	if l.logger.IsDebug() {
		if args[0] != "msg" {
			l.logger.Debug(`unexpected logging call, first key should be "msg": ` + fmt.Sprint(args...))
		}
		l.logger.Debug(fmt.Sprint(args[1]), args[2:]...)
	}
}

func mapToHclogLevel(level Level) hclog.Level {
	return hclog.Level(int(hclog.Error) - int(level))
}
