package flink

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newArtifactVersionDownloadCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "download <name>",
		Short:       "Download a Flink artifact's content in Confluent Platform.",
		Long:        "Download the binary content of a Flink artifact in Confluent Platform. Defaults to the latest version unless `--version` is specified.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactVersionDownloadOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Download the latest version of Flink artifact "my-artifact" in the environment "my-environment".`,
				Code: "confluent flink artifact version download my-artifact --output-file my-artifact.jar --environment my-environment",
			},
		),
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("output-file", "", "Path to write the downloaded artifact file.")
	cmd.Flags().String("version", "", "Version of the artifact to download. Defaults to the latest version.")
	cmd.Flags().Bool("force", false, "Overwrite the output file if it already exists.")
	addCmfFlagSet(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("output-file"))

	return cmd
}

func (c *command) artifactVersionDownloadOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	outputFile, err := cmd.Flags().GetString("output-file")
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if !force {
		if _, statErr := os.Stat(outputFile); statErr == nil {
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf(`file "%s" already exists`, outputFile),
				"Use the `--force` flag to overwrite the existing file.",
			)
		}
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	downloadedFile, err := client.DownloadArtifactContent(c.createContext(), environment, name, version)
	if err != nil {
		return err
	}
	defer os.Remove(downloadedFile.Name())
	defer downloadedFile.Close()

	destination, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, downloadedFile); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	output.Printf(false, "Downloaded Flink artifact %q to \"%s\".\n", name, outputFile)
	return nil
}
