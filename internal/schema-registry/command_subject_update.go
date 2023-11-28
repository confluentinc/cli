package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	schemaregistry "github.com/confluentinc/cli/v3/pkg/schema-registry"
)

func (c *command) newSubjectUpdateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <subject>",
		Short:   "Update subject compatibility or mode.",
		Args:    cobra.ExactArgs(1),
		RunE:    c.subjectUpdate,
		Example: examples.BuildExampleString(),
	}

	example1 := examples.Example{
		Text: `Update subject-level compatibility of subject "payments".`,
		Code: "confluent schema-registry subject update payments --compatibility backward",
	}
	example2 := examples.Example{
		Text: `Update subject-level compatibility of subject "payments" and set compatibility group to "application.version".`,
		Code: "confluent schema-registry subject update payments --compatibility backward --compatibility-group application.version",
	}
	example3 := examples.Example{
		Text: `Update subject-level mode of subject "payments".`,
		Code: "confluent schema-registry subject update payments --mode readwrite",
	}
	if cfg.IsOnPremLogin() {
		example1.Code += " " + onPremAuthenticationMsg
		example2.Code += " " + onPremAuthenticationMsg
		example3.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2, example3)

	addCompatibilityFlag(cmd)
	addCompatibilityGroupFlag(cmd)
	addMetadataDefaultsFlag(cmd)
	addMetadataOverridesFlag(cmd)
	addRulesetDefaultsFlag(cmd)
	addRulesetOverridesFlag(cmd)
	addModeFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}

	if cfg.IsCloudLogin() {
		// Deprecated
		pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

		// Deprecated
		pcmd.AddApiSecretFlag(cmd)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))
	}

	return cmd
}

func (c *command) subjectUpdate(cmd *cobra.Command, args []string) error {
	subject := args[0]

	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	compatibility, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}

	if compatibility != "" && mode != "" {
		return fmt.Errorf(errors.CompatibilityOrModeErrorMsg)
	}

	if compatibility != "" {
		return c.updateCompatibility(cmd, subject, client)
	}

	if mode != "" {
		return c.updateMode(subject, mode, client)
	}

	return fmt.Errorf(errors.CompatibilityOrModeErrorMsg)
}

func (c *command) updateCompatibility(cmd *cobra.Command, subject string, client *schemaregistry.Client) error {
	req, err := c.getConfigUpdateRequest(cmd)
	if err != nil {
		return err
	}

	if _, err := client.UpdateSubjectLevelConfig(subject, req); err != nil {
		return catchSchemaNotFoundError(err, subject, "")
	}

	output.Printf(c.Config.EnableColor, "Successfully updated subject-level compatibility to \"%s\" for subject \"%s\".\n", req.Compatibility, subject)
	return nil
}

func (c *command) updateMode(subject, mode string, client *schemaregistry.Client) error {
	res, err := client.UpdateMode(subject, srsdk.ModeUpdateRequest{Mode: srsdk.PtrString(strings.ToUpper(mode))})
	if err != nil {
		return catchSchemaNotFoundError(err, "subject", "")
	}

	output.Printf(c.Config.EnableColor, "Successfully updated subject-level mode to \"%s\" for subject \"%s\".\n", res.Mode, subject)
	return nil
}
