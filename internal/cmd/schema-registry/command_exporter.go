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
	"net/http"
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
	Name   string
	State  string
	Offset string
	Ts     string
	Trace  string
}

var (
	describeInfoLabels              = []string{"Name", "Subjects", "ContextType", "Context", "Config"}
	describeInfoHumanRenames        = map[string]string{"ContextType": "Context Type", "Config": "Remote Schema Registry Configs"}
	describeInfoStructuredRenames   = map[string]string{"Name": "name", "Subjects": "subjects", "ContextType": "context_type", "Context": "context", "Config": "config"}
	describeStatusLabels            = []string{"Name", "State", "Offset", "Ts", "Trace"}
	describeStatusHumanRenames      = map[string]string{"State": "Exporter State", "Offset": "Exporter Offset", "Ts": "Exporter Timestamp", "Trace": "Error Trace"}
	describeStatusStructuredRenames = map[string]string{"Name": "name", "State": "state", "Offset": "offset", "Ts": "timestamp", "Trace": "trace"}
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
		Use:   "create",
		Short: "Create new schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create new schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter create --name my_exporter"+
					" --subjects my_subject1,my_subject2 --context-type CUSTOM --context my_context"+
					"--config-file ~/config.txt", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	cmd.Flags().String("subjects", "", "The subjects of the exporter.")
	cmd.Flags().String("context-type", "", "The context type of the exporter.")
	cmd.Flags().String("context", "", "The context of the exporter.")
	cmd.Flags().String("config-file", "", "The file containing configurations of the exporter.")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("subjects")
	_ = cmd.MarkFlagRequired("context-type")
	_ = cmd.MarkFlagRequired("config-file")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "update",
		Short: "Update configs or information of schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update information of new schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter update --name my_exporter"+
					" --subjects my_subject1,my_subject2 --context-type CUSTOM --context my_context", cliName),
			},
			examples.Example{
				Text: "Update configs of new schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter update --config-file ~/config.txt", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("subjects", "", "The subjects of the exporter.")
	cmd.Flags().String("context-type", "", "The context type of the exporter.")
	cmd.Flags().String("context", "", "The context of the exporter.")
	cmd.Flags().String("config-file", "", "The file containing configurations of the exporter.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the information of the schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the information of schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter describe --name my_exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "get-config",
		Short: "Get the configurations of the schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.getConfig),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get the configurations of schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter get-config --name my_exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, "json", output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "status",
		Short: "Get the status of the schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.status),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get the status of schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter status --name my_exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "pause",
		Short: "Pause schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.pause),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Pause schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter pause --name my_exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "resume",
		Short: "Resume schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.resume),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Resume schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter resume --name my_exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.reset),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Reset schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter reset --name my_exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete schema exporter.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete schema exporter.",
				Code: fmt.Sprintf("%s schema-registry exporter delete --name my_exporter", cliName),
			},
		),
	}
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().String("name", "", "The name of the exporter.")
	_ = cmd.MarkFlagRequired("name")
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

func (c *exporterCommand) create(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	subjects, err := cmd.Flags().GetString("subjects")
	if err != nil {
		return err
	}
	contextType, err := cmd.Flags().GetString("context-type")
	if err != nil {
		return err
	}
	context := "."
	if contextType == "CUSTOM" {
		context, err = cmd.Flags().GetString("context")
		if err != nil {
			return err
		}
	}
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	subjectList := strings.Split(subjects, ",")
	configMap, err := readConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	_, _, err = srClient.DefaultApi.CreateExporter(ctx, srsdk.CreateExporterRequest{
		Name:        name,
		Subjects:    subjectList,
		ContextType: contextType,
		Context:     context,
		Config:      configMap,
	})
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.CreatedExporterMsg, name)
	return nil
}

func (c *exporterCommand) update(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	info, httpResponse, err := srClient.DefaultApi.GetExporterInfo(ctx, name)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return errors.Errorf(errors.SchemaExporterNotFoundMsg, name)
		}
		return err
	}

	updateRequest := srsdk.UpdateExporterRequest{
		Subjects: info.Subjects,
		ContextType: info.ContextType,
		Context: info.Context,
		Config: info.Config,
	}

	contextType, err := cmd.Flags().GetString("context-type")
	if err != nil {
		return err
	}
	if contextType != "" {
		updateRequest.ContextType = contextType
		if contextType == "CUSTOM" {
			context, err := cmd.Flags().GetString("context")
			if err != nil {
				return err
			}
			updateRequest.Context = context
		}
	}
	if cmd.Flags().Lookup("subjects").Changed {
		subjects, err := cmd.Flags().GetString("subjects")
		if err != nil {
			return err
		}
		subjectList := strings.Split(subjects, ",")
		updateRequest.Subjects = subjectList
	}
	if cmd.Flags().Lookup("config-file").Changed {
		configFile, err := cmd.Flags().GetString("config-file")
		if err != nil {
			return err
		}
		configMap, err := readConfigsFromFile(configFile)
		if err != nil {
			return err
		}
		updateRequest.Config = configMap
	}

	_, _, err = srClient.DefaultApi.PutExporter(ctx, name, updateRequest)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.UpdatedExporterMsg, name)
	return nil
}

func (c *exporterCommand) describe(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	info, httpResponse, err := srClient.DefaultApi.GetExporterInfo(ctx, name)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return errors.Errorf(errors.SchemaExporterNotFoundMsg, name)
		}
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

func (c *exporterCommand) getConfig(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	configs, httpResponse, err := srClient.DefaultApi.GetExporterConfig(ctx, name)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return errors.Errorf(errors.SchemaExporterNotFoundMsg, name)
		}
		return err
	}
	return output.StructuredOutput(outputFormat, configs)
}

func (c *exporterCommand) status(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	status, httpResponse, err := srClient.DefaultApi.GetExporterStatus(ctx, name)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return errors.Errorf(errors.SchemaExporterNotFoundMsg, name)
		}
		return err
	}

	data := &exporterStatusDisplay{
		Name:   status.Name,
		State:  status.State,
		Offset: strconv.FormatInt(status.Offset, 10),
		Ts:     strconv.FormatInt(status.Ts, 10),
		Trace:  status.Trace,
	}
	return output.DescribeObject(cmd, data, describeStatusLabels, describeStatusHumanRenames, describeStatusStructuredRenames)
}

func (c *exporterCommand) pause(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	_, httpResponse, err := srClient.DefaultApi.PauseExporter(ctx, name)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return errors.Errorf(errors.SchemaExporterNotFoundMsg, name)
		}
		return err
	}

	utils.Printf(cmd, errors.PausedExporterMsg, name)
	return nil
}

func (c *exporterCommand) resume(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	_, httpResponse, err := srClient.DefaultApi.ResumeExporter(ctx, name)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return errors.Errorf(errors.SchemaExporterNotFoundMsg, name)
		}
		return err
	}

	utils.Printf(cmd, errors.ResumedExporterMsg, name)
	return nil
}

func (c *exporterCommand) reset(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	_, httpResponse, err := srClient.DefaultApi.ResetExporter(ctx, name)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return errors.Errorf(errors.SchemaExporterNotFoundMsg, name)
		}
		return err
	}

	utils.Printf(cmd, errors.ResetExporterMsg, name)
	return nil
}

func (c *exporterCommand) delete(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	httpResponse, err := srClient.DefaultApi.DeleteExporter(ctx, name)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return errors.Errorf(errors.SchemaExporterNotFoundMsg, name)
		}
		return err
	}

	utils.Printf(cmd, errors.DeletedExporterMsg, name)
	return nil
}
