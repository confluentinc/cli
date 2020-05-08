package log

import (
	"strings"
)

/*
Sets the verbosity of the log by looking for verbosity flags.
No need to return flag validation error as cobra will handle that.
Used to allow verbosity settings of actions taken before commands are
actually executed (e.g. loading config).
 */
func SetLoggingVerbosity(args []string, logger *Logger) {
	verbosity := getVerbosity(args)
	level := getLoggerLevel(verbosity)
	logger.SetLevel(level)
}

func getVerbosity(args []string) int {
	count := 0
	for _, arg := range args {
		count += getVerbosityFlagCount(arg)
	}
	return count
}

func getVerbosityFlagCount(arg string) int {
	if strings.HasPrefix(arg, "--") {
		return getFullFlagCount(arg)
	} else if strings.HasPrefix(arg, "-") {
		return getShortHandFlagCount(arg)
	}
	return 0
}

func getFullFlagCount(arg string) int {
	if arg == "--verbose" {
		return 1
	}
	return 0
}

func getShortHandFlagCount(arg string) int {
	if isValidShortHandFlag(arg) {
		return strings.Count(arg, "v")
	}
	return 0
}

func isValidShortHandFlag(arg string) bool {
	if !strings.HasPrefix(arg, "-") {
		return false
	}
	if len(arg) == 1 {
		return false
	}
	flagValue := arg[1:]
	return !strings.Contains(flagValue, "-")
}



func getLoggerLevel(verbosity int) Level {
	switch verbosity {
	case 0:
		return ERROR
	case 1:
		return WARN
	case 2:
		return INFO
	case 3:
		return DEBUG
	case 4:
		return TRACE
	default:
		// if more than 4 we give max level
		return TRACE
	}
}
