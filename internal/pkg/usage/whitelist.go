package usage

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/types"
)

func WhitelistCommandsAndFlags(cmd *cobra.Command, whitelist types.Set[string]) {
	whitelist.Add(cmd.Name())
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		whitelist.Add(flag.Name)
	})
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		whitelist.Add(flag.Name)
	})

	for _, subcommand := range cmd.Commands() {
		WhitelistCommandsAndFlags(subcommand, whitelist)
	}
}
