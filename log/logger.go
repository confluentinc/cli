package log

import (
	"io"
	"os"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/sirupsen/logrus"
)

// Logger is the standard logger for the Confluent SDK.
type Logger struct {
	l *logrus.Logger
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

// New create and configures a new Logger.
func New() *Logger {
	logger := &Logger{l: logrus.New()}
	logger.l.Formatter = &logrus.TextFormatter{FullTimestamp: true, DisableLevelTruncation: true}
	logger.SetLevel(WARN)
	logger.SetOutput(os.Stdout)
	return logger
}

func (l *Logger) Debug(args ...interface{}) {
	l.l.Debug(args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.l.Debugf(format, args...)
}


func (l *Logger) Info(args ...interface{}) {
	l.l.Info(args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.l.Infof(format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.l.Warn(args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.l.Warnf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.l.Error(args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.l.Errorf(format, args...)
}

// Log logs a "msg" and key-value pairs.
// Example: Log("msg", "hello", "key1", "val1", "key2", "val2")
func (l *Logger) Log(args ...interface{}) {
	var msg interface{}
	m := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		k := args[i].(string)
		v := args[i+1]
		if k == "msg" {
			msg = v
		} else {
			m[k] = v
		}
	}
	l.l.WithFields(logrus.Fields(m)).Debug(msg)
}

func (l *Logger) SetLevel(level Level) {
	switch level {
	case ERROR:
		l.l.SetLevel(logrus.ErrorLevel)
	case WARN:
		l.l.SetLevel(logrus.WarnLevel)
	case INFO:
		l.l.SetLevel(logrus.InfoLevel)
	case DEBUG:
		l.l.SetLevel(logrus.DebugLevel)
	case TRACE:
		l.l.SetLevel(logrus.TraceLevel)
	}
}

func (l *Logger) SetOutput(out io.Writer) {
	l.l.SetOutput(out)
}
