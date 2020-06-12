package feedback

import (
	"fmt"
	"os"
)

var (
	feedbackNudge = "Did you know you can use the \"ccloud feedback\" command to send the team feedback?\nLet us know if the ccloud CLI is meeting your needs, or what we can do to improve it."
)

func HandleFeedbackNudge(cliName string, cmdArgs []string, cmdErr error) {
	if cliName == "ccloud" && (cmdErr != nil || isHumanReadableOrHelp(cmdArgs)) {
		_, _ = fmt.Fprintln(os.Stderr, feedbackNudge)
	}
}

func isHumanReadableOrHelp(args []string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-o" || args[i] == "--output" {
			return args[i+1] == "human"
		}
		if args[i] == "-h" || args[i] == "--help" {
			return true
		}
	}
	return true
}
