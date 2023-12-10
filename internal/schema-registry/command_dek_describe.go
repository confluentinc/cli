package schemaregistry

import (
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDekDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a Dek.",
		Args:  cobra.NoArgs,
		RunE:  c.dekDescribe,
	}

	cmd.Flags().String("name", "", "Name of the KEK.")
	cmd.Flags().String("subject", "", "Subject of the DEK.")
	pcmd.AddAlgorithmFlag(cmd)
	cmd.Flags().String("version", "", "Version of the Dek.")
	cmd.Flags().Bool("deleted", false, "Include deleted DEK.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) dekDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	if !cmd.Flags().Changed("subject") {
		subjects, err := client.GetDekSubjects(name)
		if err != nil {
			return err
		}
		list := output.NewList(cmd)
		for _, subject := range subjects {
			list.Add(&subjectListOut{Subject: subject})
		}
		return list.Print()
	}

	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	algorithm, err := cmd.Flags().GetString("algorithm")
	if err != nil {
		return err
	}

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	var getErr error
	var dek srsdk.Dek
	if version != "" {
		dek, getErr = client.GetDekByVersion(name, subject, version, algorithm, deleted)
	} else {
		versions, err := client.GetDeKVersions(name, subject, algorithm, deleted)
		if err != nil {
			return err
		}
		list := output.NewList(cmd)
		for _, version := range versions {
			list.Add(&versionOut{Version: version})
		}
		err = list.Print()
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			return nil
		}

		dek, getErr = client.GetDek(name, subject, algorithm, deleted)
	}

	if getErr != nil {
		return getErr
	}

	return printDek(cmd, dek)
}
