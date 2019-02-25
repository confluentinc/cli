package common

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/log"
)

func SetLoggingVerbosity(cmd *cobra.Command, logger *log.Logger) error {
	verbosity, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		return err
	}
	switch verbosity {
	case 1:
		logger.SetLevel(logrus.InfoLevel)
	case 2:
		logger.SetLevel(logrus.DebugLevel)
	case 3:
		logger.SetLevel(logrus.TraceLevel)
	}
	return nil
}
