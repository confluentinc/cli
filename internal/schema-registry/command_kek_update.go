package schemaregistry

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newKekUpdateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a Kek.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kekUpdate,
	}

	// all descriptions need to be updated. @RobertY
	cmd.Flags().StringSlice("kms-props", nil, "A comma-separated list?")
	cmd.Flags().String("doc", "", "")
	cmd.Flags().Bool("shared", false, "Share the KEK.") // ?

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd) // guess it's needed?
	}
	pcmd.AddOutputFlag(cmd) // ? hmm?

	return cmd
}

func (c *command) kekUpdate(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	kek, err := client.DescribeKek(args[0], false)
	if err != nil {
		return err
	}

	updateReq := srsdk.UpdateKekRequest{
		KmsProps: kek.KmsProps,
		Doc:      kek.Doc,
		Shared:   kek.Shared,
	}

	if cmd.Flags().Changed("kms-props") {
		kmsPropsSlices, err := cmd.Flags().GetStringSlice("kms-props")
		if err != nil {
			return err
		}

		kmsProps := make(map[string]string)
		for _, item := range kmsPropsSlices {
			pair := strings.Split(item, ":")
			if len(pair) != 2 {
				return errors.New("ill format") // updated this...
			}
			kmsProps[pair[0]] = pair[1]
		}
		updateReq.SetKmsProps(kmsProps)
	}

	if cmd.Flags().Changed("doc") {
		doc, err := cmd.Flags().GetString("doc")
		if err != nil {
			return err
		}
		updateReq.SetDoc(doc)
	}

	if cmd.Flags().Changed("shared") {
		shared, err := cmd.Flags().GetBool("shared")
		if err != nil {
			return err
		}
		updateReq.SetShared(shared)
	}

	res, err := client.UpdateKek(args[0], updateReq)
	if err != nil {
		return err
	}

	return printKek(cmd, res)
}
