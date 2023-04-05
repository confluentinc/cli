package prompt

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/go-ps1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

const none = "(none)"

func New(cfg *v1.Config) *cobra.Command {
	prompt := ps1.New(pversion.CLIName, []ps1.Token{
		{
			Name: 'a',
			Desc: `The current Kafka API key in use. E.g., "ABCDEF1234567890"`,
			Func: func() string {
				ctx := cfg.Context()
				if ctx == nil {
					return none
				}
				kcc := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
				if kcc == nil || kcc.APIKey == "" {
					return none
				}
				return kcc.APIKey
			},
		},
		{
			Name: 'C',
			Desc: `The name of the current context in use. E.g., "dev-app1", "stag-dc1", "prod"`,
			Func: func() string {
				if cfg.CurrentContext == "" {
					return none
				}
				return utils.CropString(cfg.CurrentContext, 30)
			},
		},
		{
			Name: 'e',
			Desc: `The ID of the current environment in use. E.g., "a-4567"`,
			Func: func() string {
				ctx := cfg.Context()
				if ctx == nil {
					return none
				}
				if id := ctx.GetCurrentEnvironment(); id != "" {
					return id
				}
				return none
			},
		},
		{
			Name: 'k',
			Desc: `The ID of the current Kafka cluster in use. E.g., "lkc-abc123"`,
			Func: func() string {
				ctx := cfg.Context()
				if ctx == nil {
					return none
				}
				kcc := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
				if kcc == nil {
					return none
				} else {
					return kcc.ID
				}
			},
		},
		{
			Name: 'K',
			Desc: `The name of the current Kafka cluster in use. E.g., "prod-us-west-2-iot"`,
			Func: func() string {
				ctx := cfg.Context()
				if ctx == nil {
					return none
				}
				kcc := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
				if kcc == nil || kcc.Name == "" {
					return none
				}
				return kcc.Name
			},
		},
		{
			Name: 'u',
			Desc: `The current user or credentials in use. E.g., "joe@montana.com"`,
			Func: func() string {
				ctx := cfg.Context()
				if ctx == nil {
					return none
				}
				if email := ctx.GetUser().GetEmail(); email != "" {
					return email
				}
				return none
			},
		},
	})

	cmd := prompt.NewCobraCommand()
	cmd.Short = fmt.Sprintf("Add %s context to your terminal prompt.", pversion.FullCLIName)

	cmd.ResetFlags()
	cmd.Flags().StringP("format", "f", "(confluent|%C)", "The format string to use. See the help for details.")
	cmd.Flags().IntP("timeout", "t", 200, "The maximum execution time in milliseconds.")

	return cmd
}
