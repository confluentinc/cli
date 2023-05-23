package controller

import (
	"log"
	"os"

	"golang.org/x/term"
)

func getStdin() *term.State {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		log.Println("Couldn't get stdin state with term.GetState" + err.Error())
		return nil
	}
	return state
}

func restoreStdin(state *term.State) {
	if state != nil {
		_ = term.Restore(int(os.Stdin.Fd()), state)
	}
}
