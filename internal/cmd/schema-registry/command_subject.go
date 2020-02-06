package schema_registry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/go-printer"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type subjectCommand struct {
	*pcmd.AuthenticatedCLICommand
	srClient *srsdk.APIClient
}

// NewSubjectCommand returns the Cobra command for Schema Registry subject list
func NewSubjectCommand(config *v2.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "subject",
			Short: "Manage Schema Registry subjects.",
		},
		config, prerunner)
	subjectCmd := &subjectCommand{
		AuthenticatedCLICommand: cliCmd,
		srClient:                srClient,
	}
	subjectCmd.init()
	return subjectCmd.Command
}

func (c *subjectCommand) init() {

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List subjects.",
		Example: FormatDescription(`
Retrieve all subjects available in a Schema Registry

::
		config.CLIName schema-registry subject list
`, c.Config.CLIName),
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	c.AddCommand(listCmd)
	// Update
	updateCmd := &cobra.Command{
		Use:   "update <subjectname> [--compatibility <compatibility>] [--mode <mode>] ",
		Short: "Update subject compatibility or mode.",
		Example: FormatDescription(`
Update subject level compatibility or mode of schema registry.

::
		config.CLIName schema-registry subject update <subjectname> --compatibility=BACKWARD
		config.CLIName schema-registry subject update <subjectname> --mode=READWRITE
`, c.Config.CLIName),
		RunE: c.update,
		Args: cobra.ExactArgs(1),
	}
	updateCmd.Flags().String("compatibility", "", "Can be BACKWARD, BACKWARD_TRANSITIVE, FORWARD, FORWARD_TRANSITIVE, FULL, FULL_TRANSITIVE, or NONE.")
	updateCmd.Flags().String("mode", "", "Can be READWRITE, READ, OR WRITE.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	// Describe
	describeCmd := &cobra.Command{
		Use:   "describe <subjectname>",
		Short: "Describe subject versions and compatibility.",
		Example: FormatDescription(`
Retrieve all versions registered under a given subject and its compatibility level.

::
		config.CLIName schema-registry subject describe <subjectname>
`, c.Config.CLIName),
		RunE: c.describe,
		Args: cobra.ExactArgs(1),
	}
	c.AddCommand(describeCmd)
}

func (c *subjectCommand) update(cmd *cobra.Command, args []string) error {
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if compat != "" {
		return c.updateCompatibility(cmd, args)
	}
	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if mode != "" {
		return c.updateMode(cmd, args)
	}
	return errors.New("flag --compatibility or --mode is required.")
}
func (c *subjectCommand) updateCompatibility(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	updateReq := srsdk.ConfigUpdateRequest{Compatibility: compat}
	_, _, err = srClient.DefaultApi.UpdateSubjectLevelConfig(ctx, args[0], updateReq)
	if err != nil {
		return err
	}
	pcmd.Println(cmd, "Successfully updated")
	return nil
}

func (c *subjectCommand) updateMode(cmd *cobra.Command, args []string) error {
	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	updatedMode, _, err := srClient.DefaultApi.UpdateMode(ctx, args[0], srsdk.ModeUpdateRequest{Mode: mode})
	if err != nil {
		return err
	}
	pcmd.Println(cmd, "Successfully updated Subject level Mode: "+updatedMode.Mode)
	return nil
}

func (c *subjectCommand) list(cmd *cobra.Command, args []string) error {
	var listLabels = []string{"Subject"}
	var data [][]string
	type listDisplay struct {
		Subject string
	}
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {

		return err
	}
	list, _, err := srClient.DefaultApi.List(ctx)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		for _, l := range list {
			data = append(data, printer.ToRow(&listDisplay{
				Subject: l,
			}, listLabels))
		}
		printer.RenderCollectionTable(data, listLabels)
	} else {
		pcmd.Println(cmd, "No subjects")
	}
	return nil
}

func (c *subjectCommand) describe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	versions, _, err := srClient.DefaultApi.ListVersions(ctx, args[0])
	if err != nil {
		return err
	}
	PrintVersions(versions)
	return nil
}
