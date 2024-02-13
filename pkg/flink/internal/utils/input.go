package utils

import (
	"os"

	"golang.org/x/term"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/log"
)

func GetStdin() *term.State {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		log.CliLogger.Warnf("Couldn't get stdin state with term.GetState: %v", err)
		return nil
	}
	return state
}

func GetConsoleParser() (prompt.ConsoleParser, error) {
	consoleParser, err := prompt.NewStandardInputParser()
	if err != nil {
		log.CliLogger.Warnf("Couldn't create console parser: %v", err)
		return nil, err
	}
	if err := consoleParser.Setup(); err != nil {
		log.CliLogger.Warnf("Couldn't setup console parser: %v", err)
		return nil, err
	}
	return consoleParser, nil
}

func TearDownConsoleParser(consoleParser prompt.ConsoleParser) {
	err := consoleParser.TearDown()
	if err != nil {
		log.CliLogger.Warnf("Couldn't tear down console parser: %v", err)
	}
}

func RestoreStdin(state *term.State) {
	if state != nil {
		_ = term.Restore(int(os.Stdin.Fd()), state)
	}
}
