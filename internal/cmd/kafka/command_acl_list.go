package kafka

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (c *aclCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}

	cmd.Flags().AddFlagSet(resourceFlags())
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
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

	resourceIdMap, err := c.mapUserIdToResourceId()
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST != nil {
		opts := aclBindingToClustersClusterIdAclsGetOpts(acl[0].ACLBinding)

		kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		aclGetResp, httpResp, err := kafkaREST.Client.ACLApi.ClustersClusterIdAclsGet(kafkaREST.Context, lkc, &opts)

		if err != nil && httpResp != nil {
			// Kafka REST is available, but an error occurred
			return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
			// Kafka REST is available and there was no error
			return aclutil.PrintACLsFromKafkaRestResponseWithResourceIdMap(cmd, aclGetResp, cmd.OutOrStdout(), resourceIdMap)
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := pcmd.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	resp, err := c.Client.Kafka.ListACLs(context.Background(), cluster, convertToFilter(acl[0].ACLBinding))
	if err != nil {
		return err
	}

	return aclutil.PrintACLsWithResourceIdMap(cmd, resp, os.Stdout, resourceIdMap)
}
