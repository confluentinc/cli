package common

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/log"
)

func SetLoggingVerbosity(cmd *cobra.Command, logger *log.Logger) error {
	verbosity, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		return err
	}
	switch verbosity {
	case 0:
		logger.SetLevel(log.ERROR)
	case 1:
		logger.SetLevel(log.WARN)
	case 2:
		logger.SetLevel(log.INFO)
	case 3:
		logger.SetLevel(log.DEBUG)
	case 4:
		logger.SetLevel(log.TRACE)
	default:
		// requested more than 4 -v's, so let's give them the max verbosity we support
		logger.SetLevel(log.TRACE)
	}
	return nil
}
