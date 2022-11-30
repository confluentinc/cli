package ksql

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a ksqlDB cluster.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
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

	// Check KSQL exists
	cluster, err := c.V2Client.DescribeKsqlCluster(id, c.EnvironmentId())
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, id)
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.KsqlCluster, id, cluster.Spec.GetDisplayName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, cluster.Spec.GetDisplayName()); err != nil {
		return err
	}

	// When deleting a cluster we need to remove all the associated topics. This operation will succeed only if cluster
	// is UP and provisioning didn't fail. If provisioning failed we can't connect to the ksql server, so we can't delete
	// the topics.
	if c.getClusterStatus(&cluster) == "PROVISIONED" {
		if err := c.deleteTopics(cluster.GetId(), cluster.Status.GetHttpEndpoint()); err != nil {
			return err
		}
	}

	if err := c.V2Client.DeleteKsqlCluster(id, c.EnvironmentId()); err != nil {
		return err
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.KsqlCluster, id)
	return nil
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
	//this returns a 503
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
