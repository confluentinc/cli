package schemaregistry

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

type exporterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

type exporterInfoDisplay struct {
	Name        string
	Subjects    string
	ContextType string
	Context     string
	Config      string
}

type exporterStatusDisplay struct {
	Name      string
	State     string
	Offset    string
	Timestamp string
	Trace     string
}

var (
	describeInfoLabels              = []string{"Name", "Subjects", "ContextType", "Context", "Config"}
	describeInfoHumanRenames        = map[string]string{"ContextType": "Context Type", "Config": "Remote Schema Registry Configs"}
	describeInfoStructuredRenames   = map[string]string{"Name": "name", "Subjects": "subjects", "ContextType": "context_type", "Context": "context", "Config": "config"}
	describeStatusLabels            = []string{"Name", "State", "Offset", "Timestamp", "Trace"}
	describeStatusHumanRenames      = map[string]string{"State": "Exporter State", "Offset": "Exporter Offset", "Timestamp": "Exporter Timestamp", "Trace": "Error Trace"}
	describeStatusStructuredRenames = map[string]string{"Name": "name", "State": "state", "Offset": "offset", "Timestamp": "timestamp", "Trace": "trace"}
)

func NewExporterCommand(cliName string, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:   "exporter",
			Short: "Manage Schema Registry exporters.",
		}, prerunner, ExporterSubcommandFlags)
	exporterCmd := &exporterCommand{
		AuthenticatedStateFlagCommand: cliCmd,
		srClient:                      srClient,
	}
	exporterCmd.init(cliName)
	return exporterCmd.Command
}

func (c *exporterCommand) init(cliName string) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all schema exporters.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all schema exporters.",
				Code: fmt.Sprintf("%s schema-registry exporter list", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "create <name>",
		Short: "Create new schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create new schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter create my-exporter --subjects my_subject1,my_subject2 --context-type CUSTOM --context my_context --config config.txt", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().StringSlice("subjects", []string{"*"}, "Exporter subjects. Use a comma separated list, or specify the flag multiple times.")
	cmd.Flags().String("context-type", "AUTO", `Exporter context type. One of "AUTO", "CUSTOM" or "NONE".`)
	cmd.Flags().String("context-name", "", "Exporter context name.")
	cmd.Flags().String("config-file", "", "Exporter config file.")

	_ = cmd.MarkFlagRequired("config-file")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "update <name>",
		Short: "Update configs or information of schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update information of new schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter update my-exporter"+
					" --subjects my_subject1,my_subject2 --context-type CUSTOM --context-name my_context", cliName),
			},
			examples.Example{
				Text: "Update configs of new schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter update my-exporter --config-file ~/config.txt", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().StringSlice("subjects", []string{}, "The subjects of the exporter. Should use"+
		" comma separated list, or specify the flag multiple times.")
	cmd.Flags().String("context-type", "", `The context type of the exporter. Can be "AUTO", "CUSTOM" or "NONE".`)
	cmd.Flags().String("context-name", "", "The context name of the exporter.")
	cmd.Flags().String("config-file", "", "The file containing configurations of the exporter.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the information of the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the information of schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter describe my-exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "get-config <name>",
		Short: "Get the configurations of the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.getConfig),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get the configurations of schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter get-config my-exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, "json", output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "get-status <name>",
		Short: "Get the status of the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.status),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get the status of schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter get-status my-exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "pause <name>",
		Short: "Pause schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.pause),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Pause schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter pause my-exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "resume <name>",
		Short: "Resume schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.resume),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Resume schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter resume my-exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "reset <name>",
		Short: "Reset schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.reset),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Reset schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter reset my-exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter delete my-exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *exporterCommand) list(cmd *cobra.Command, _ []string) error {
	type listDisplay struct {
		Exporter string
	}

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	exporters, _, err := srClient.DefaultApi.GetExporters(ctx)
	if err != nil {
		return err
	}

	if len(exporters) > 0 {
		outputWriter, err := output.NewListOutputWriter(cmd, []string{"Exporter"}, []string{"Exporter"}, []string{"Exporter"})
		if err != nil {
			return err
		}
		for _, exporter := range exporters {
			outputWriter.AddElement(&listDisplay{
				Exporter: exporter,
			})
		}
		return outputWriter.Out()
	} else {
		utils.Println(cmd, errors.NoExporterMsg)
	}
	return nil
}

func (c *exporterCommand) create(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]
	subjects, err := cmd.Flags().GetStringSlice("subjects")
	if err != nil {
		return err
	}
	contextType, err := cmd.Flags().GetString("context-type")
	if err != nil {
		return err
	}
	context := "."
	if contextType == "CUSTOM" {
		context, err = cmd.Flags().GetString("context-name")
		if err != nil {
			return err
		}
	} else if cmd.Flags().Changed("context-name") {
		return errors.New("can only set context-name if context-type is CUSTOM")
	}
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	configMap, err := utils.ReadConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	_, _, err = srClient.DefaultApi.CreateExporter(ctx, srsdk.CreateExporterRequest{
		Name:        name,
		Subjects:    subjects,
		ContextType: contextType,
		Context:     context,
		Config:      configMap,
	})
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Created", name)
	return nil
}

func (c *exporterCommand) update(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]
	info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, name)
	if err != nil {
		return err
	}

	updateRequest := srsdk.UpdateExporterRequest{
		Subjects:    info.Subjects,
		ContextType: info.ContextType,
		Context:     info.Context,
		Config:      nil,
	}

	contextType, err := cmd.Flags().GetString("context-type")
	if err != nil {
		return err
	}
	if contextType != "" {
		updateRequest.ContextType = contextType
	}
	context, err := cmd.Flags().GetString("context-name")
	if err != nil {
		return err
	}
	if context != "" {
		updateRequest.Context = context
	}
	subjects, err := cmd.Flags().GetStringSlice("subjects")
	if err != nil {
		return err
	}
	if len(subjects) > 0 {
		updateRequest.Subjects = subjects
	}
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	if configFile != "" {
		configMap, err := utils.ReadConfigsFromFile(configFile)
		if err != nil {
			return err
		}
		updateRequest.Config = configMap
	}

	_, _, err = srClient.DefaultApi.PutExporter(ctx, name, updateRequest)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Updated", name)
	return nil
}

func (c *exporterCommand) describe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]
	info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, name)
	if err != nil {
		return err
	}

	data := &exporterInfoDisplay{
		Name:        info.Name,
		Subjects:    strings.Join(info.Subjects, ", "),
		ContextType: info.ContextType,
		Context:     info.Context,
		Config:      convertMapToString(info.Config),
	}
	return output.DescribeObject(cmd, data, describeInfoLabels, describeInfoHumanRenames, describeInfoStructuredRenames)
}

func (c *exporterCommand) getConfig(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	configs, _, err := srClient.DefaultApi.GetExporterConfig(ctx, name)
	if err != nil {
		return err
	}
	return output.StructuredOutputForCommand(cmd, outputFormat, configs)
}

func (c *exporterCommand) status(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]
	status, _, err := srClient.DefaultApi.GetExporterStatus(ctx, name)
	if err != nil {
		return err
	}

	data := &exporterStatusDisplay{
		Name:      status.Name,
		State:     status.State,
		Offset:    strconv.FormatInt(status.Offset, 10),
		Timestamp: strconv.FormatInt(status.Ts, 10),
		Trace:     status.Trace,
	}
	return output.DescribeObject(cmd, data, describeStatusLabels, describeStatusHumanRenames, describeStatusStructuredRenames)
}

func (c *exporterCommand) pause(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]

	_, _, err = srClient.DefaultApi.PauseExporter(ctx, name)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Paused", name)
	return nil
}

func (c *exporterCommand) resume(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]

	_, _, err = srClient.DefaultApi.ResumeExporter(ctx, name)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Resumed", name)
	return nil
}

func (c *exporterCommand) reset(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]

	_, _, err = srClient.DefaultApi.ResetExporter(ctx, name)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Reset", name)
	return nil
}

func (c *exporterCommand) delete(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name := args[0]

	_, err = srClient.DefaultApi.DeleteExporter(ctx, name)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Deleted", name)
	return nil
}
