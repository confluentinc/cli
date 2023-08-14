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
		log.CliLogger.Warnf("Couldn't get stdin state with term.GetState. Error: %v", err)
		return nil
	}
	return state
}

func GetConsoleParser() prompt.ConsoleParser {
	if fileInfo, err := os.Stat("/dev/tty"); err != nil {
		log.CliLogger.Warnf(`Couldn't open "/dev/tty" file. Error: %v`, err)
		return nil
	} else if fileInfo.Mode().Perm()&0444 == 0 {
		log.CliLogger.Warn(`Couldn't read "/dev/tty" file because read permissions are not set.`)
		return nil
	}
	consoleParser := prompt.NewStandardInputParser()
	err := consoleParser.Setup()
	if err != nil {
		log.CliLogger.Warnf("Couldn't setup console parser. Error: %v", err)
	}
	return consoleParser
}

func TearDownConsoleParser(consoleParser prompt.ConsoleParser) {
	err := consoleParser.TearDown()
	if err != nil {
		log.CliLogger.Warnf("Couldn't tear down console parser. Error: %v", err)
	}
}

func RestoreStdin(state *term.State) {
	if state != nil {
		_ = term.Restore(int(os.Stdin.Fd()), state)
	}
}
