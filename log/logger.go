package log

import (
	"github.com/sirupsen/logrus"
)

// Logger is the standard logger for the Confluent SDK.
type Logger struct {
	*logrus.Logger
}

// New create and configures a new Logger.
func New() *Logger {
	logger := &Logger{Logger: logrus.New()}
	logger.Formatter = &logrus.TextFormatter{FullTimestamp: true, DisableLevelTruncation: true}
	return logger
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
	l.WithFields(logrus.Fields(m)).Debug(msg)
}
