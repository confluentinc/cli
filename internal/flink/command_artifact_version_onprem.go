package flink

import (
	"github.com/spf13/cobra"
)

func (c *command) newArtifactVersionCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage versions of a Flink artifact in Confluent Platform.",
	}

	cmd.AddCommand(c.newArtifactVersionCreateCommandOnPrem())
	cmd.AddCommand(c.newArtifactVersionDeleteCommandOnPrem())
	cmd.AddCommand(c.newArtifactVersionDescribeCommandOnPrem())
	cmd.AddCommand(c.newArtifactVersionDownloadCommandOnPrem())
	cmd.AddCommand(c.newArtifactVersionListCommandOnPrem())

	return cmd
}
