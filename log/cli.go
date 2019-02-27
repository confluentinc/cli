package log

import "github.com/spf13/cobra"

func SetLoggingVerbosity(cmd *cobra.Command, logger *Logger) error {
	verbosity, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		return err
	}
	switch verbosity {
	case 0:
		logger.SetLevel(WARN)
	case 1:
		logger.SetLevel(INFO)
	case 2:
		logger.SetLevel(DEBUG)
	case 3:
		logger.SetLevel(TRACE)
	default:
		// requested more than 3 -v's, so let's give them the max verbosity we support
		logger.SetLevel(TRACE)
	}
	return nil
}
