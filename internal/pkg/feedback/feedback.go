package feedback

import (
	"fmt"
	"os"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func HandleFeedbackNudge(args []string) {
	if isHelpCommand(args) {
		_, _ = fmt.Fprintf(os.Stderr, errors.FeedbackNudgeMsg)
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
