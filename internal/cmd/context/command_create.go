package context

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <context>",
		Short: "Create a new context.",
		Long:  "Create a new Cloud context with an API key.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a context called "new context":`,
				Code: version.CLIName + ` context create "new context" --bootstrap https://example.com --api-key key --api-secret @api-secret.txt`,
			},
		),
	}

	cmd.Flags().String("bootstrap", "", "Bootstrap URL.")
	cmd.Flags().String("api-key", "", "API key.")
	cmd.Flags().String("api-secret", "", "API secret. Can be specified as plaintext, as a file, starting with '@', or as stdin, starting with '-'.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("bootstrap"))
	cobra.CheckErr(cmd.MarkFlagRequired("api-key"))
	cobra.CheckErr(cmd.MarkFlagRequired("api-secret"))

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	bootstrap, err := c.parseStringFlag(cmd, "bootstrap", "Bootstrap URL: ", false)
	if err != nil {
		return err
	}

	apiKey, err := c.parseStringFlag(cmd, "api-key", "API Key: ", false)
	if err != nil {
		return err
	}

	apiSecret, err := c.parseStringFlag(cmd, "api-secret", "API Secret: ", true)
	if err != nil {
		return err
	}

	name := args[0]

	if err := c.Config.CreateContext(name, bootstrap, apiKey, apiSecret); err != nil {
		return err
	}

	ctx, err := c.Config.FindContext(name)
	if err != nil {
		return err
	}

	return describeContext(cmd, ctx)
}

// parseStringFlag gets the value of a flag, potentially via an interactive prompt.
func (c *command) parseStringFlag(cmd *cobra.Command, name, prompt string, secure bool) (string, error) {
	str, err := cmd.Flags().GetString(name)
	if err != nil {
		return "", err
	}

	val, err := c.resolver.ValueFrom(str, prompt, secure)
	if err != nil {
		return "", err
	}

	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return "", errors.Errorf(errors.CannotBeEmptyErrorMsg, name)
	}

	return val, nil
}
