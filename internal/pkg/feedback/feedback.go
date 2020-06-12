package feedback

import (
	"fmt"
	"os"
)

var (
	feedbackNudge = "\nDid you know you can use the \"ccloud feedback\" command to send the team feedback?\nLet us know if the ccloud CLI is meeting your needs, or what we can do to improve it."
)

func HandleFeedbackNudge(cliName string, cmdArgs []string, cmdErr error) {
	if cliName == "ccloud" &&  isHelpOrIsCommandFailureInHumanReadableMode(cmdArgs, cmdErr) {
		_, _ = fmt.Fprintln(os.Stderr, feedbackNudge)
	}
}

func isHelpOrIsCommandFailureInHumanReadableMode(args []string, err error) bool {
	for i := 0; i < len(args); i++ {
		if isHelp(args[i]) {
			return true
		}
		if i < len(args) - 1 {
			if err != nil && isHumanReadable(args[i], args[i+1]) {
				return true
			}
		}
	}
	return false
}

func isHelp(flag string) bool {
	return flag == "-h" || flag == "--help"
}

func isHumanReadable(flag string, flagVlaue string) bool {
	return (flag == "-o" || flag == "--output") && flagVlaue == "human"
}
