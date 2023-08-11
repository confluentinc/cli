package utils

import (
	"os"

	"golang.org/x/term"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/log"
)

func GetStdin() *term.State {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		log.CliLogger.Warnf("Couldn't get stdin state with term.GetState. Error: %v\n", err)
		return nil
	}
	return state
}

func GetConsoleParser() prompt.ConsoleParser {
	if fileInfo, err := os.Stat("/dev/tty"); err != nil {
		log.CliLogger.Warnf("Couldn't open \"/dev/tty\" file because it doesn't exist. Error: %v\n", err)
	} else if fileInfo.Mode().Perm()&0444 != 0 {
		// Checks if any read permissions are set for "/dev/tty" file
		log.CliLogger.Warnf("Error: No read permissions are not set for \"/dev/tty\".")
	}
	consoleParser := prompt.NewStandardInputParser()
	err := consoleParser.Setup()
	if err != nil {
		log.CliLogger.Warnf("Couldn't setup console parser. Error: %v\n", err)
	}
	return consoleParser
}

func TearDownConsoleParser(consoleParser prompt.ConsoleParser) {
	err := consoleParser.TearDown()
	if err != nil {
		log.CliLogger.Warnf("Couldn't tear down console parser. Error: %v\n", err)
	}
}

func RestoreStdin(state *term.State) {
	if state != nil {
		_ = term.Restore(int(os.Stdin.Fd()), state)
	}
}
