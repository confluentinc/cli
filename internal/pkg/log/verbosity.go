package log

import (
	"strings"
)


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
		if arg == "--verbose" {
			return 1
		}
	} else if strings.HasPrefix(arg, "-") {
		return strings.Count(arg, "v")
	}
	return 0
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
