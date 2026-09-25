package flink

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newArtifactVersionDownloadCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <name>",
		Short: "Download a Flink artifact's content in Confluent Platform.",
		Long:  "Download the binary content of a Flink artifact in Confluent Platform. Defaults to the latest version unless `--version` is specified.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.artifactVersionDownloadOnPrem,
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

	// Like `asyncapi export`, download overwrites the output file if it already exists.
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	downloadedFile, err := client.DownloadArtifactContent(c.createContext(), environment, name, version)
	if err != nil {
		return err
	}
	// A 2xx response with an empty body yields a nil file (the SDK does not allocate one); guard against it so the
	// deferred cleanup below doesn't dereference a nil *os.File.
	if downloadedFile == nil {
		return fmt.Errorf(`no content was returned for Flink artifact "%s"`, name)
	}
	defer os.Remove(downloadedFile.Name())
	defer downloadedFile.Close()

	if err := replaceFile(outputFile, downloadedFile); err != nil {
		return err
	}

	output.Printf(false, "Downloaded Flink artifact %q to %q.\n", name, outputFile)
	return nil
}

// replaceFile writes content to a temporary file in path's directory and renames it over path, so a failed write
// never truncates or corrupts a file already at path. An existing file keeps its permissions; a new one gets 0644.
func replaceFile(path string, content io.Reader) error {
	perm := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}

	temp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	// Removes the temporary file if any step below fails; after a successful rename it no longer exists.
	defer os.Remove(temp.Name())

	if _, err := io.Copy(temp, content); err != nil {
		temp.Close()
		return fmt.Errorf("failed to write output file: %w", err)
	}
	if err := temp.Chmod(perm); err != nil {
		temp.Close()
		return fmt.Errorf("failed to write output file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	return nil
}
