package connector
//
//import (
//	"context"
//
//	"github.com/spf13/cobra"
//
//	ccsdk "github.com/confluentinc/ccloud-sdk-go"
//	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
//	"github.com/confluentinc/cli/internal/pkg/config"
//	"github.com/confluentinc/cli/internal/pkg/errors"
//	"github.com/confluentinc/cli/internal/pkg/log"
//)
//
//type taskCommand struct {
//	*cobra.Command
//	config       *config.Config
//	ccClient     ccsdk.Connect
//	ch           *pcmd.ConfigHelper
//	logger       *log.Logger
//}
//
//func NewTaskCommand(config *config.Config, ccloudClient ccsdk.Connect, ch *pcmd.ConfigHelper, logger *log.Logger) *cobra.Command {
//	taskCmd := &taskCommand{
//		Command: &cobra.Command{
//			Use:   "task",
//			Short: "Manage Connector tasks.",
//		},
//		config:       config,
//		ccClient:     ccloudClient,
//		ch:           ch,
//		logger:       logger,
//	}
//	taskCmd.init()
//	return taskCmd.Command
//}
//
//func (c *taskCommand) init() {
//	createCmd := &cobra.Command{
//		Use:     "list",
//		Short:   `Enable Schema Registry for this environment.`,
//		Example: FormatDescription(`{{.CLIName}} connector task list --connector-id <connector-id>`, c.config.CLIName),
//		RunE:    c.list,
//		Args:    cobra.MaximumNArgs(1),
//	}
//	createCmd.Flags().String("connector-id", "", "Connector ID.")
//	_ = createCmd.MarkFlagRequired("connector-id")
//	c.AddCommand(createCmd)
//}
//
//func (c *taskCommand) list(cmd *cobra.Command, args []string) error {
//	ctx := context.Background()
//	// Collect the parameters
//	accountId, err := pcmd.GetEnvironment(cmd, c.config)
//	if err != nil {
//		return errors.HandleCommon(err, cmd)
//	}
//	//connectorID, err := cmd.Flags().GetString("connector-id")
//	//if err != nil {
//	//	return errors.HandleCommon(err, cmd)
//	//}
//	 return nil
//}
