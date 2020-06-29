package feedback

import (
	"fmt"
	"os"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func HandleFeedbackNudge(cliName string, cmdArgs []string) {
	if cliName == "ccloud" && isHelpCommand(cmdArgs) {
		_, _ = fmt.Fprintln(os.Stderr, errors.FeedbackNudgeMsg)
	}
}

func isHelpCommand(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}
