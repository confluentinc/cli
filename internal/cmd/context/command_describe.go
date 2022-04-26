package context

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [context]",
		Short:             "Describe a context.",
		Long:              "Describe a context or a specific context field.",
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	cmd.Flags().Bool("api-key", false, "Get the API key for a context.")
	cmd.Flags().Bool("username", false, "Get the username for a context.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	ctx, err := c.context(args)
	if err != nil {
		return err
	}
	credential := ctx.Credential

	apiKey, err := cmd.Flags().GetBool("api-key")
	if err != nil {
		return err
	}
	if apiKey && credential.CredentialType != v1.APIKey {
		return fmt.Errorf(`context "%s" does not have an associated API key`, ctx.Name)
	}

	username, err := cmd.Flags().GetBool("username")
	if err != nil {
		return err
	}
	if username && credential.CredentialType != v1.Username {
		return fmt.Errorf(`context "%s" does not have an associated username`, ctx.Name)
	}

	if apiKey {
		utils.Println(cmd, credential.APIKeyPair.Key)
		return nil
	}

	if username {
		utils.Println(cmd, credential.Username)
		return nil
	}

	return describeContext(cmd, ctx)
}

func describeContext(cmd *cobra.Command, ctx *dynamicconfig.DynamicContext) error {
	var (
		listFields        = []string{"Name", "Platform", "Credential"}
		humanRenames      = map[string]string{"Name": "Name", "Platform": "Platform", "Credential": "Credential"}
		structuredRenames = map[string]string{"Name": "name", "Platform": "platform", "Credential": "credential"}
	)

	row := &struct {
		Name       string
		Platform   string
		Credential string
	}{
		Name:       ctx.Name,
		Platform:   ctx.PlatformName,
		Credential: ctx.CredentialName,
	}

	return output.DescribeObject(cmd, row, listFields, humanRenames, structuredRenames)
}
