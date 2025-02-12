package iam

import (
	"github.com/spf13/cobra"

	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type certificatePoolCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type certificatePoolOut struct {
	Id                 string `human:"ID" serialized:"id"`
	Name               string `human:"Name" serialized:"name"`
	Description        string `human:"Description" serialized:"description"`
	ExternalIdentifier string `human:"External Identifier" serialized:"external_identifier"`
	Filter             string `human:"Filter" serialized:"filter"`
}

func newCertificatePoolCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "certificate-pool",
		Short:       "Manage certificate pools.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Hidden:      !(cfg.IsTest || featureflags.Manager.BoolVariation("cli.mtls", cfg.Context(), config.CliLaunchDarklyClient, true, false)),
	}

	c := &certificatePoolCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand(cfg))
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printCertificatePool(cmd *cobra.Command, certificatePool certificateauthorityv2.IamV2CertificateIdentityPool) error {
	table := output.NewTable(cmd)
	table.Add(&certificatePoolOut{
		Id:                 certificatePool.GetId(),
		Name:               certificatePool.GetDisplayName(),
		Description:        certificatePool.GetDescription(),
		ExternalIdentifier: certificatePool.GetExternalIdentifier(),
		Filter:             certificatePool.GetFilter(),
	})
	return table.Print()
}

func (c *certificatePoolCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return nil
	}

	return pcmd.AutocompleteCertificatePool(c.V2Client, provider)
}

func (c *certificatePoolCommand) AddProviderFlag(cmd *cobra.Command) {
	cmd.Flags().String("provider", "", "ID of this pool's certificate authority.")

	pcmd.RegisterFlagCompletionFunc(cmd, "provider", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return pcmd.AutocompleteCertificateAuthorities(c.V2Client)
	})
}
