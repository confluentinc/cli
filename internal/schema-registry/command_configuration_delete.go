package schemaregistry

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newConfigurationDeleteCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete top-level or subject-level schema configuration.",
		Args:  cobra.NoArgs,
		RunE:  c.configDelete,
	}

	example1 := examples.Example{
		Text: "Delete the top-level configuration.",
		Code: "confluent schema-registry configuration delete",
	}
	example2 := examples.Example{
		Text: `Delete the subject-level configuration of subject "payments".`,
		Code: "confluent schema-registry configuration delete --subject payments",
	}
	if cfg.IsOnPremLogin() {
		example1.Code += " " + onPremAuthenticationMsg
		example2.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	cmd.Flags().String("subject", "", subjectUsage)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationAndClientPathFlags(cmd)
	}
	addSchemaRegistryEndpointFlag(cmd)
	pcmd.AddOutputFlag(cmd)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) configDelete(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	var outStr string
	if subject != "" {
		promptMsg := fmt.Sprintf(`Are you sure you want to delete the subject-level compatibility level config and revert it to the global default for "%s"?`, subject)
		if err := deletion.ConfirmPrompt(cmd, promptMsg); err != nil {
			return err
		}
		outStr, err = client.DeleteSubjectLevelConfig(subject)
		if err != nil {
			return catchSubjectLevelConfigNotFoundError(err, subject)
		}
	} else {
		promptMsg := `Are you sure you want to delete the global compatibility level config and revert it to the default?`
		if err := deletion.ConfirmPrompt(cmd, promptMsg); err != nil {
			return err
		}
		outStr, err = client.DeleteTopLevelConfig()
		if err != nil {
			return err
		}
	}

	output.Printf(c.Config.EnableColor, "Deleted %s.\n", resource.SchemaRegistryConfiguration)
	out := &configOut{}
	if err := json.Unmarshal([]byte(outStr), out); err != nil {
		return err
	}
	table := output.NewTable(cmd)
	table.Add(out)
	return table.PrintWithAutoWrap(false)
}
