package usage

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/set"
)

func WhitelistCommandsAndFlags(cmd *cobra.Command, whitelist set.Set) {
	whitelist.Add(cmd.Name())
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		whitelist.Add(flag.Name)
	})

	for _, subcommand := range cmd.Commands() {
		WhitelistCommandsAndFlags(subcommand, whitelist)
	}
}
