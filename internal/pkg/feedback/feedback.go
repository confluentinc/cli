package feedback

import (
	"fmt"
	"os"
)

var (
	feedbackNudge = "\nDid you know you can use the \"ccloud feedback\" command to send the team feedback?\nLet us know if the ccloud CLI is meeting your needs, or what we can do to improve it."
)

func HandleFeedbackNudge(cliName string, cmdArgs []string, cmdErr error) {
	if cliName == "ccloud" &&  isHelpCommand(cmdArgs, cmdErr) {
		_, _ = fmt.Fprintln(os.Stderr, feedbackNudge)
	}
}

func isHelpCommand(args []string, err error) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}
