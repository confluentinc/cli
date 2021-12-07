package output

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type output int

func (o output) String() string {
	return validFlagValues[o]
}

const (
	Human output = iota
	JSON
	YAML
)

const FlagName = "output"

var validFlagValues = []string{"human", "json", "yaml"}

func AddFlag(cmd *cobra.Command) {
	AddFlagWithDefaultValue(cmd, Human.String())
}

func AddFlagWithDefaultValue(cmd *cobra.Command, defaultValue string) {
	cmd.Flags().StringP(FlagName, "o", defaultValue, `Specify the output format as "human", "json", or "yaml".`)

	pcmd.RegisterFlagCompletionFunc(cmd, FlagName, func(_ *cobra.Command, _ []string) []string {
		return validFlagValues
	})
}
