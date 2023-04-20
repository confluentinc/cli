package ksql

import (
	"context"
	"io"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *ksqlCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete ksqlDB clusters.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *ksqlCommand) delete(cmd *cobra.Command, args []string) error {
	id := args[0]
	log.CliLogger.Debugf("Deleting ksqlDB cluster \"%v\".\n", id)

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	idToCluster := make(map[string]ksqlv2.KsqldbcmV2Cluster)
	if err := c.confirmDeletion(cmd, environmentId, args, idToCluster); err != nil {
		return err
	}

	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deleted []string
	for _, id := range args {
		// When deleting a cluster we need to remove all the associated topics. This operation will succeed only if cluster
		// is UP and provisioning didn't fail. If provisioning failed we can't connect to the ksql server, so we can't delete
		// the topics.
		cluster := idToCluster[id]
		if c.getClusterStatus(&cluster) == "PROVISIONED" {
			if err := c.deleteTopics(cluster.GetId(), cluster.Status.GetHttpEndpoint()); err != nil {
				errs = multierror.Append(errs, err)
				continue
			}
		}

		if err := c.V2Client.DeleteKsqlCluster(id, environmentId); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessMsg(deleted, resource.KsqlCluster)

	return errs.ErrorOrNil()
}

func (c *ksqlCommand) deleteTopics(clusterId, endpoint string) error {
	ctx := c.Config.Context()
	state, err := ctx.AuthenticatedState()
	if err != nil {
		return err
	}

	bearerToken, err := pauth.GetBearerToken(state, ctx.Platform.Server, clusterId)
	if err != nil {
		return err
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: bearerToken})
	client := sling.New().Client(oauth2.NewClient(context.Background(), ts)).Base(endpoint)
	request := map[string][]string{"deleteTopicList": {".*"}}
	response, err := client.Post("/ksql/terminate").BodyJSON(&request).ReceiveSuccess(nil)
	// this returns a 503
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return errors.Errorf(errors.KsqlDBTerminateClusterErrorMsg, clusterId, string(body))
	}
	return nil
}

func (c *ksqlCommand) confirmDeletion(cmd *cobra.Command, environmentId string, args []string, idToCluster map[string]ksqlv2.KsqldbcmV2Cluster) error {
	describeFunc := func(id string) error {
		cluster, err := c.V2Client.DescribeKsqlCluster(id, environmentId)
		if err == nil {
			idToCluster[id] = cluster
		}
		return err
	}

	if err := deletion.ValidateArgs(cmd, args, resource.KsqlCluster, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.KsqlCluster, args[0], idToCluster[args[0]].Spec.GetDisplayName()); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.KsqlCluster, args); err != nil || !ok {
			return err
		}
	}

	return nil
}
