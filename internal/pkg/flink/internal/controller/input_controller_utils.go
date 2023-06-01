package controller

import (
	"os"

	"golang.org/x/term"

	"github.com/confluentinc/cli/internal/pkg/log"
)

func getStdin() *term.State {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		log.CliLogger.Warnf("Couldn't get stdin state with term.GetState. Error: %v \n", err)
		return nil
	}
	return state
}

func restoreStdin(state *term.State) {
	if state != nil {
		_ = term.Restore(int(os.Stdin.Fd()), state)
	}
}
