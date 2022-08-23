package ksql

import (
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/log"
	"io/ioutil"
	"net/http"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newDeleteCommand(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             fmt.Sprintf("Delete a ksqlDB %s.", resource),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *ksqlCommand) delete(cmd *cobra.Command, args []string) error {
	id := args[0]
	log.CliLogger.Debugf("Deleting cluster: %v", id)

	req := &schedv1.KSQLCluster{
		AccountId: c.EnvironmentId(),
		Id:        id,
	}

	// Check KSQL exists
	cluster, err := c.Client.KSQL.Describe(context.Background(), req)
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, id)
	}

	// When deleting a cluster we need to remove all the associated topics. This operation will succeed only if cluster
	// is UP and provisioning didn't fail. If provisioning failed we can't connect to the ksql server, so we can't delete
	// the topics.
	if cluster.Status == schedv1.ClusterStatus_UP {
		provisioningFailed, err := c.checkProvisioningFailed(cluster)
		if !provisioningFailed && err != nil {
			if err := c.deleteTopics(req); err != nil {
				return err
			}
		}
	}

	if err := c.Client.KSQL.Delete(context.Background(), req); err != nil {
		return err
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.KsqlCluster, args[0])
	return nil
}

func (c *ksqlCommand) deleteTopics(cluster *schedv1.KSQLCluster) error {
	ctx := c.Config.Context()
	state, err := ctx.AuthenticatedState()
	if err != nil {
		return err
	}

	bearerToken, err := pauth.GetBearerToken(state, ctx.Platform.Server, cluster.Id)
	if err != nil {
		return err
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: bearerToken})
	client := sling.New().Client(oauth2.NewClient(context.Background(), ts)).Base(cluster.Endpoint)
	request := make(map[string][]string)
	request["deleteTopicList"] = []string{".*"}
	response, err := client.Post("/ksql/terminate").BodyJSON(&request).ReceiveSuccess(nil)
	//this returns a 503
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return errors.Errorf(errors.KsqlDBTerminateClusterErrorMsg, cluster.Id, string(body))
	}
	return nil
}
