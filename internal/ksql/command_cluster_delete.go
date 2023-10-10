package ksql

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"

	pauth "github.com/confluentinc/cli/v3/pkg/auth"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *ksqlCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more ksqlDB clusters.",
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
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	idToCluster := c.mapIdToCluster(args, environmentId)

	existenceFunc := func(id string) bool {
		_, ok := idToCluster[id]
		return ok
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.KsqlCluster, idToCluster[args[0]].Spec.GetDisplayName()); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		// When deleting a cluster we need to remove all the associated topics. This operation will succeed only if cluster
		// is UP and provisioning didn't fail. If provisioning failed we can't connect to the ksql server, so we can't delete
		// the topics.
		cluster := idToCluster[id]
		if c.getClusterStatus(&cluster) == "PROVISIONED" {
			if err := c.deleteTopics(cluster.GetId(), cluster.Status.GetHttpEndpoint()); err != nil {
				return err
			}
		}

		return c.V2Client.DeleteKsqlCluster(id, environmentId)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.KsqlCluster)
	return err
}

func (c *ksqlCommand) deleteTopics(clusterId, endpoint string) error {
	ctx := c.Config.Context()
	state, err := ctx.AuthenticatedState()
	if err != nil {
		return err
	}

	dataplaneToken, err := pauth.GetDataplaneToken(state, ctx.Platform.Server)
	if err != nil {
		return err
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: dataplaneToken})

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
		return fmt.Errorf(`failed to terminate ksqlDB cluster "%s" due to "%s"`, clusterId, string(body))
	}
	return nil
}

func (c *ksqlCommand) mapIdToCluster(args []string, environmentId string) map[string]ksqlv2.KsqldbcmV2Cluster {
	// NOTE: This function does not return an error for invalid IDs; validation will instead
	// be done by deletion.ValidateAndConfirmDeletion using this map. This allows for consistent existence
	// error messaging across all delete commands which support multiple deletion.

	idToCluster := make(map[string]ksqlv2.KsqldbcmV2Cluster)
	for _, id := range args {
		cluster, err := c.V2Client.DescribeKsqlCluster(id, environmentId)
		if err == nil {
			idToCluster[id] = cluster
		}
	}

	return idToCluster
}
