package kafka

import (
	"context"
	"fmt"
	"net/http"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *aclCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  c.delete,
	}

	cmd.Flags().AddFlagSet(aclConfigFlags())
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *aclCommand) delete(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}

	userIdMap, err := c.mapResourceIdToUserId()
	if err != nil {
		return err
	}

	if err := c.aclResourceIdToNumericId(acls, userIdMap); err != nil {
		return err
	}

	var filters []*schedv1.ACLFilter
	for _, acl := range acls {
		validateAddAndDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		filters = append(filters, convertToFilter(acl.ACLBinding))
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST != nil {
		kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		kafkaRestURL := kafkaREST.Client.GetConfig().BasePath

		kafkaRestExists := true
		matchingBindingCount := 0
		for i, filter := range filters {
			deleteOpts := aclFilterToClustersClusterIdAclsDeleteOpts(filter)
			deleteResp, httpResp, err := kafkaREST.Client.ACLV3Api.DeleteKafkaAcls(kafkaREST.Context, lkc, &deleteOpts)

			if err != nil && httpResp == nil {
				if i == 0 {
					// Kafka REST is not available, fallback to KafkaAPI
					kafkaRestExists = false
					break
				}
				// i > 0: unlikely
				printAclsDeleted(cmd, matchingBindingCount)
				return kafkaRestError(kafkaRestURL, err, httpResp)
			}

			if err != nil {
				if i > 0 {
					// unlikely
					printAclsDeleted(cmd, matchingBindingCount)
				}
				return kafkaRestError(kafkaRestURL, err, httpResp)
			}

			if httpResp.StatusCode == http.StatusOK {
				matchingBindingCount += len(deleteResp.Data)
			} else {
				printAclsDeleted(cmd, matchingBindingCount)
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
		}

		if kafkaRestExists {
			// Kafka REST is available and at least one ACL was deleted
			printAclsDeleted(cmd, matchingBindingCount)
			return nil
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	matchingBindingCount := 0
	for _, acl := range acls {
		// For the tests it's useful to know that the ListACLs call is coming from the delete call.
		ctx := context.WithValue(context.Background(), kafka.Requester, "delete")

		resp, err := c.Client.Kafka.ListACLs(ctx, cluster, convertToFilter(acl.ACLBinding))
		if err != nil {
			return err
		}
		matchingBindingCount += len(resp)
	}
	if matchingBindingCount == 0 {
		utils.ErrPrintf(cmd, errors.ACLsNotFoundMsg)
		return nil
	}

	if err := c.Client.Kafka.DeleteACLs(context.Background(), cluster, filters); err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.DeletedACLsMsg)
	return nil
}

func printAclsDeleted(cmd *cobra.Command, count int) {
	if count == 0 {
		utils.ErrPrintf(cmd, errors.ACLsNotFoundMsg)
	} else {
		utils.ErrPrintf(cmd, fmt.Sprintf(errors.DeletedACLsCountMsg, count))
	}
}
