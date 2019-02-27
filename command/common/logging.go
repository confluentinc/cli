package common

import (
	"github.com/confluentinc/cli/log"
	"github.com/spf13/cobra"
)

func SetLoggingVerbosity(cmd *cobra.Command, logger *log.Logger) error {
	verbosity, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		return err
	}
	switch verbosity {
	case 0:
		logger.SetLevel(log.WARN)
	case 1:
		logger.SetLevel(log.INFO)
	case 2:
		logger.SetLevel(log.DEBUG)
	case 3:
		logger.SetLevel(log.TRACE)
	default:
		// requested more than 3 -v's, so let's give them the max verbosity we support
		logger.SetLevel(log.TRACE)
	}
	return nil
}
