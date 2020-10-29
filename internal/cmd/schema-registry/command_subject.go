package schema_registry

import (
	"fmt"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type subjectCommand struct {
	*pcmd.AuthenticatedCLICommand
	srClient *srsdk.APIClient
}

// NewSubjectCommand returns the Cobra command for Schema Registry subject list
func NewSubjectCommand(cliName string, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "subject",
			Short: "Manage Schema Registry subjects.",
		}, prerunner)
	subjectCmd := &subjectCommand{
		AuthenticatedCLICommand: cliCmd,
		srClient:                srClient,
	}
	subjectCmd.init(cliName)
	return subjectCmd.Command
}

func (c *subjectCommand) init(cliName string) {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List subjects.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Retrieve all subjects available in a Schema Registry:",
				Code: fmt.Sprintf("%s schema-registry subject list", cliName),
			},
		),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().BoolP("deleted", "D", false, "View the deleted subjects.")
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	updateCmd := &cobra.Command{
		Use:   "update <subject-name> [--compatibility <compatibility>] [--mode <mode>]",
		Short: "Update subject compatibility or mode.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update subject level compatibility or mode of Schema Registry:",
				Code: fmt.Sprintf("%s schema-registry subject update <subject-name> --compatibility=BACKWARD\n%s schema-registry subject update <subject-name> --mode=READWRITE", cliName, cliName),
			},
		),
	}
	updateCmd.Flags().String("compatibility", "", "Can be BACKWARD, BACKWARD_TRANSITIVE, FORWARD, FORWARD_TRANSITIVE, FULL, FULL_TRANSITIVE, or NONE.")
	updateCmd.Flags().String("mode", "", "Can be READWRITE, READ, OR WRITE.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <subject-name>",
		Short: "Describe subject versions and compatibility.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Retrieve all versions registered under a given subject and its compatibility level.",
				Code: fmt.Sprintf("%s schema-registry subject describe <subject-name>", cliName),
			},
		),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().BoolP("deleted", "D", false, "View the deleted schema.")
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)
}

func (c *subjectCommand) update(cmd *cobra.Command, args []string) error {
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	if compat != "" {
		return c.updateCompatibility(cmd, args)
	}
	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	if mode != "" {
		return c.updateMode(cmd, args)
	}
	return errors.New(errors.CompatibilityOrModeErrorMsg)
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
	utils.Printf(cmd, errors.UpdatedSubjectLevelCompatibilityMsg, compat, args[0])
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
	utils.Printf(cmd, errors.UpdatedSubjectLevelModeMsg, updatedMode, args[0])
	return nil
}

func (c *subjectCommand) list(cmd *cobra.Command, _ []string) error {
	type listDisplay struct {
		Subject string
	}
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}
	listOpts := srsdk.ListOpts{Deleted: optional.NewBool(deleted)}
	list, _, err := srClient.DefaultApi.List(ctx, &listOpts)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		outputWriter, err := output.NewListOutputWriter(cmd, []string{"Subject"}, []string{"Subject"}, []string{"subject"})
		if err != nil {
			return err
		}
		for _, l := range list {
			outputWriter.AddElement(&listDisplay{
				Subject: l,
			})
		}
		return outputWriter.Out()
	} else {
		utils.Println(cmd, errors.NoSubjectsMsg)
	}
	return nil
}

func (c *subjectCommand) describe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}
	listVersionsOpts := srsdk.ListVersionsOpts{Deleted: optional.NewBool(deleted)}
	versions, _, err := srClient.DefaultApi.ListVersions(ctx, args[0], &listVersionsOpts)
	if err != nil {
		return err
	}
	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if outputOption == output.Human.String() {
		PrintVersions(versions)
	} else {
		structuredOutput := &struct {
			Version []int32
		}{
			versions,
		}
		fields := []string{"Version"}
		structuredRenames := map[string]string{"Version": "version"}
		return output.DescribeObject(cmd, structuredOutput, fields, map[string]string{}, structuredRenames)
	}
	return nil
}
