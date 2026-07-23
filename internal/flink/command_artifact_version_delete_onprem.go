package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newArtifactVersionDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete <name>",
		Short:       "Delete a version of a Flink artifact in Confluent Platform.",
		Long:        "Delete a version of a Flink artifact in Confluent Platform. Deleting the last remaining version deletes the artifact.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactVersionDeleteOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete version 2 of Flink artifact "my-artifact" in the environment "my-environment".`,
				Code: "confluent flink artifact version delete my-artifact --version 2 --environment my-environment",
			},
		),
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("version", "", `Version of the artifact to delete, or "all" to delete every version.`)
	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))

	return cmd
}

func (c *command) artifactVersionDeleteOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	if _, err := client.DescribeArtifact(c.createContext(), environment, name, ""); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), artifactLookupSuggestions)
	}

	var promptMsg string
	if version == "all" {
		promptMsg = fmt.Sprintf(`Are you sure you want to delete all versions of Flink artifact "%s"?`, name)
	} else {
		promptMsg = fmt.Sprintf(`Are you sure you want to delete version "%s" of Flink artifact "%s"?`, version, name)
	}
	if err := deletion.ConfirmPrompt(cmd, promptMsg); err != nil {
		return err
	}

	if err := client.DeleteArtifact(c.createContext(), environment, name, version); err != nil {
		return err
	}

	if version == "all" {
		output.Printf(false, "Deleted all versions of Flink artifact %q.\n", name)
	} else {
		output.Printf(false, "Deleted version %q of Flink artifact %q.\n", version, name)
	}
	return nil
}
