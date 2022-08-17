package kafka

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (c *aclCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().AddFlagSet(resourceFlags())
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("principal", "", `Principal for this operation, prefixed with "User:".`)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	acl, err := parse(cmd)
	if err != nil {
		return err
	}

	userIdMap, err := c.mapResourceIdToUserId()
	if err != nil {
		return err
	}

	if err := c.aclResourceIdToNumericId(acl, userIdMap); err != nil {
		return err
	}

	if acl[0].errors != nil {
		return acl[0].errors
	}

	resourceIdMap, err := c.mapUserIdToResourceId()
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST != nil {
		kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
		if err != nil {
			return err
		}

		aclGetResp, httpResp, err := kafkaREST.CloudClient.GetKafkaAcls(kafkaClusterConfig.ID, acl[0].ACLBinding)
		if err != nil && httpResp != nil {
			// Kafka REST is available, but an error occurred
			return kafkaRestError(kafkaREST.CloudClient.GetKafkaRestUrl(), err, httpResp)
		}
		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
			// Kafka REST is available and there was no error
			return aclutil.PrintACLsFromKafkaRestResponseWithResourceIdMap(cmd, aclGetResp, cmd.OutOrStdout(), resourceIdMap)
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	resp, err := c.Client.Kafka.ListACLs(context.Background(), cluster, convertToFilter(acl[0].ACLBinding))
	if err != nil {
		return err
	}

	return aclutil.PrintACLsWithResourceIdMap(cmd, resp, os.Stdout, resourceIdMap)
}
