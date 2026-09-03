package log

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/hashicorp/go-hclog"
	"github.com/mattn/go-isatty"
)

// VerbosityEnvVar sets the log level when no -v flag is passed, using the same 0-4 scale as -v.
// A flag always wins, since -v is a count flag and 0 can only mean "not passed". Level 5
// (UNSAFE_TRACE) is unreachable here; it stays gated behind the --unsafe-trace flag.
const VerbosityEnvVar = "CONFLUENT_VERBOSITY"

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
	out    io.Writer
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
	// UNSAFE_TRACE is for printing sensitive information such as HTTP requests and responses
	UNSAFE_TRACE
)

// New creates and configures a new Logger
func New(level Level, output io.Writer) *Logger {
	return &Logger{
		Level: level,
		logger: hclog.New(&hclog.LoggerOptions{
			Output:          output,
			Level:           mapToHclogLevel(level),
			Color:           colorForOutput(output),
			ColorHeaderOnly: true,
		}),
		out: output,
	}
}

// hclog's AutoColor leaves color on for any writer without a file descriptor, which would put
// escape codes into test buffers, so decide explicitly rather than delegating.
func colorForOutput(output io.Writer) hclog.ColorOption {
	file, ok := output.(*os.File)
	if !ok {
		return hclog.ColorOff
	}

	if isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd()) {
		return hclog.ForceColor
	}
	return hclog.ColorOff
}

func (l *Logger) SetVerbosity(verbosity int) {
	if verbosity == 0 {
		verbosity = l.verbosityFromEnv()
	}

	level := min(Level(verbosity), UNSAFE_TRACE)

	l.Level = level
	l.logger.SetLevel(mapToHclogLevel(level))
}

// verbosityFromEnv reads the verbosity from VerbosityEnvVar. An unset variable returns 0 silently; a
// value outside 0-4 (non-integer, negative, or >= 5) returns 0 but warns, since someone who set it
// meant to raise verbosity and would otherwise get no output and no hint why. Level 5 (UNSAFE_TRACE)
// is excluded on purpose: it logs credentials, so it must come from --unsafe-trace, not an env var.
func (l *Logger) verbosityFromEnv() int {
	raw, ok := os.LookupEnv(VerbosityEnvVar)
	if !ok || raw == "" {
		return 0
	}

	verbosity, err := strconv.Atoi(raw)
	if err != nil || verbosity < 0 || verbosity > int(TRACE) {
		fmt.Fprintf(l.out, "[WARN] Ignoring invalid environment variable %q=%q; expected a number from 0 (quietest) to 4 (most verbose).\n", VerbosityEnvVar, raw)
		return 0
	}
	return verbosity
}

func (l *Logger) UnsafeTrace(args ...any) {
	message := fmt.Sprint(args...)
	// HACK: hclog.NoLevel = 0, which corresponds to UNSAFE_TRACE
	if l.logger.GetLevel() == hclog.NoLevel {
		l.logger.Trace(message)
	} else {
		l.append(UNSAFE_TRACE, message)
	}
}

func (l *Logger) UnsafeTracef(format string, args ...any) {
	l.UnsafeTrace(fmt.Sprintf(format, args...))
}

func (l *Logger) Trace(args ...any) {
	message := fmt.Sprint(args...)
	if l.Level >= TRACE { // Avoid l.logger.IsTrace() since it only checks "== TRACE" so it will miss UNSAFE_TRACE
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
		// Higher levels are more verbose, so a buffered message is emitted only if the level it
		// was logged at is now within the threshold.
		if lm.level > l.Level {
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
		case UNSAFE_TRACE:
			l.UnsafeTrace(lm.message)
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
