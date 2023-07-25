package schemaregistry

import (
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	schemaregistry "github.com/confluentinc/cli/internal/pkg/schema-registry"
)

func (c *command) newSubjectUpdateCommand(cfg *v1.Config) *cobra.Command {
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
		example1.Code += " " + OnPremAuthenticationMsg
		example2.Code += " " + OnPremAuthenticationMsg
		example3.Code += " " + OnPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2, example3)

	addCompatibilityFlag(cmd)
	addCompatibilityGroupFlag(cmd)
	addMetadataDefaultsFlag(cmd)
	addMetadataOverridesFlag(cmd)
	addRulesetDefaultsFlag(cmd)
	addRulesetOverridesFlag(cmd)
	addModeFlag(cmd)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)

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

	client, err := c.GetSchemaRegistryClient()
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
		return errors.New(errors.CompatibilityOrModeErrorMsg)
	}

	if compatibility != "" {
		return c.updateCompatibility(cmd, subject, compatibility, client)
	}

	if mode != "" {
		return c.updateMode(subject, mode, client)
	}

	return errors.New(errors.CompatibilityOrModeErrorMsg)
}

func (c *command) updateCompatibility(cmd *cobra.Command, subject, compatibility string, client *schemaregistry.Client) error {
	req, err := c.getConfigUpdateRequest(cmd)
	if err != nil {
		return err
	}

	if _, err := client.UpdateSubjectLevelConfig(subject, req); err != nil {
		return catchSchemaNotFoundError(err, subject, "")
	}

	output.Printf("Successfully updated subject level compatibility to \"%s\" for subject \"%s\".\n", compatibility, subject)
	return nil
}

func (c *command) updateMode(subject, mode string, client *schemaregistry.Client) error {
	updatedMode, err := client.UpdateMode(subject, srsdk.ModeUpdateRequest{Mode: strings.ToUpper(mode)})
	if err != nil {
		return catchSchemaNotFoundError(err, "subject", "")
	}

	output.Printf("Successfully updated subject level mode to \"%s\" for subject \"%s\".\n", updatedMode, subject)
	return nil
}
